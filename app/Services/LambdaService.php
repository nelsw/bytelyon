<?php

namespace App\Services;

use App\Models\Proxy;
use Aws\Lambda\LambdaClient;
use Illuminate\Container\Attributes\Singleton;
use Illuminate\Support\Facades\Log;

#[Singleton]
readonly class LambdaService
{
    private LambdaClient $client;

    public function __construct()
    {
        $this->client = new LambdaClient(config('services.lambda'));
    }

    private function invoke(string $functionName, array $input = []): array
    {
        $result = $this->client->invoke([
            'FunctionName' => $functionName,
            'InvocationType' => 'RequestResponse',
            'Payload' => json_encode($input),
        ]);

        $output = json_decode($result->get('Payload')->getContents(), true);

        Log::debug('LambdaService#invoke', [
            'ƒ' => $functionName,
            'in' => $input,
            'out' => $output,
        ]);

        return $output;
    }

    public function news(array $urls): array
    {
        return $this->invoke('bytelyon-news-scraper', compact('urls'));
    }

    public function page(string $url, bool $includeLinks = true): array
    {
        return $this->invoke('bytelyon-page-scraper', compact('url', 'includeLinks'));
    }

    public function grab(string $url): array
    {
        return $this->invoke('bytelyon-grab', [
            'url' => $url,
            'goto_timeout_ms' => 10_000,
        ]);
    }

    public function serp(string $query, Proxy $proxy): array
    {
        return $this->invoke('bytelyon-serp-scraper', [
            'query' => $query,
            'proxy' => [
                'protocol' => $proxy->scheme,
                'server' => $proxy->host,
                'port' => $proxy->port,
                'username' => $proxy->user,
                'password' => $proxy->pass,
                'bypass' => $proxy->bypass,
            ],
            'geoip' => true,
        ]);
    }
}
