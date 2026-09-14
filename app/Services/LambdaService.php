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

    private function invoke(string $functionName, array $input = []): array
    {
        $result = $this->client->invoke([
            'FunctionName' => $functionName,
            'InvocationType' => self::invocationType,
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
        return $this->invoke('bytelyon-serp-scraper', [
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
        ]);
    }
}
