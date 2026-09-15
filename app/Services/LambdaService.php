<?php

namespace App\Services;

use App\Data\Lambda\Payload;
use App\Models\Proxy;
use Aws\Lambda\LambdaClient;
use Illuminate\Container\Attributes\Singleton;
use Illuminate\Support\Facades\Log;

#[Singleton]
readonly class LambdaService
{
    private const string invocationType = 'RequestResponse';

    private LambdaClient $client;

    public function __construct()
    {
        $this->client = new LambdaClient(config('services.lambda'));
    }

    public function scrape(string $url): Payload
    {
        return Payload::make($this->client->invoke([
            'FunctionName' => 'bytelyon-grab',
            'InvocationType' => self::invocationType,
            'Payload' => json_encode([
                'url' => $url,
                'goto_timeout_ms' => 10_000,
            ]),
        ]));
    }

    public function serp(string $query, Proxy $proxy): array
    {
        // The Lambda handler's `proxy` field must be either a full proxy URL
        // string or a Playwright `ProxySettings`-shaped dict: {"server":
        // "<scheme>://<host>:<port>", "username", "password", "bypass"}.
        // A dict is passed straight through to Chromium unchanged (no further
        // normalizing), so `server` must already combine the scheme/host/port
        // — sending them as separate keys silently produces an invalid proxy
        // config that Chromium can't use, and the whole invocation hangs until
        // the Lambda function times out instead of failing fast.
        $server = "{$proxy->scheme}://{$proxy->host}";
        if ($proxy->port) {
            $server .= ":{$proxy->port}";
        }

        // This handler's entire purpose is routing the Google search request
        // through the proxy — `bypass` matching google would instead send that
        // request directly from Lambda's own datacenter IP, straight into the
        // IP-reputation block this proxy exists to avoid. Strip it defensively
        // regardless of what's stored on the Proxy record. Matches "google"
        // generally (not just the literal "google.com") — a stored value of
        // e.g. ".google" (missing ".com") is still clearly meant to target
        // Google and must not slip through a stricter substring check.
        $bypass = collect(explode(',', (string) $proxy->bypass))
            ->map(fn (string $host) => trim($host))
            ->reject(fn (string $host) => $host === '' || str_contains(strtolower($host), 'google'))
            ->implode(',');

        return json_decode($this->client->invoke([
            'FunctionName' => 'bytelyon-serp-scraper',
            'InvocationType' => self::invocationType,
            'Payload' => json_encode([
                'query' => $query,
                'proxy' => [
                    'server' => $server,
                    'username' => $this->proxyUsername($proxy),
                    'password' => $proxy->pass,
                    'bypass' => $bypass !== '' ? $bypass : null,
                ],
                'geoip' => true,
            ]),
        ])->get('Payload')->getContents(), true);
    }

    /**
     * Oxylabs' targeting suffixes (`-cc-<country>`, `-sessid-<id>`) require
     * the `customer-` username prefix to take effect at all — confirmed
     * empirically: the bare username silently accepts them as part of the
     * username string but the proxy responds 407, while `customer-<username>`
     * with the same suffixes works. This only normalizes that static prefix;
     * it deliberately does *not* append a per-call `-sessid-` here — one
     * `serp()` call maps to one Lambda invocation, which internally retries
     * with a fresh browser/proxy connection on failure (see
     * docker/lambda/serp-scraper/lambda_handler.py), and a sticky session id
     * fixed for the whole invocation would pin every retry to the *same*
     * exit IP, defeating the retry loop's entire point of getting a fresh one
     * per attempt. That per-attempt session id is generated Python-side
     * instead, where each retry attempt is actually visible. No-op for other
     * providers (e.g. DataImpulse's `__cr.us` convention uses incompatible
     * syntax and must not get a `customer-` prefix).
     */
    private function proxyUsername(Proxy $proxy): ?string
    {
        if (! $proxy->username || ! str_contains(strtolower($proxy->host), 'oxylabs')) {
            return $proxy->username;
        }

        return str_starts_with($proxy->username, 'customer-')
            ? $proxy->username
            : "customer-{$proxy->username}";
    }
}
