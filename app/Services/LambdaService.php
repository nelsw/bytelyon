<?php

namespace App\Services;

use App\Data\Lambda\Payload;
use App\Models\Proxy;
use Aws\Lambda\LambdaClient;
use Illuminate\Container\Attributes\Singleton;

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
        return json_decode($this->client->invoke([
            'FunctionName' => 'bytelyon-serp-scraper',
            'InvocationType' => self::invocationType,
            'Payload' => json_encode([
                'query' => $query,
                'proxy' => [
                    'protocol' => $proxy->scheme,
                    'server' => $proxy->host,
                    'port' => $proxy->port,
                    'username' => $proxy->username,
                    'password' => $proxy->pass,
                    'bypass' => $proxy->bypass,
                ],
                'geoip' => true,
            ]),
        ])->get('Payload')->getContents(), true);
    }
}
