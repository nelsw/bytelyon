<?php

namespace App\Services;

use App\Models\Proxy;
use Aws\Lambda\LambdaClient;
use Illuminate\Container\Attributes\Singleton;

#[Singleton]
readonly class LambdaService
{
    private LambdaClient $client;

    public function __construct()
    {
        $this->client = new LambdaClient(config('lambda'));
    }

    private function invoke(string $functionName, array $payload = []): array
    {
        $result = $this->client->invoke([
            'FunctionName' => $functionName,
            'InvocationType' => 'RequestResponse', // Use 'Event' for asynchronous execution
            'Payload' => json_encode($payload),
        ]);
        return json_decode($result->get('Payload')->getContents(), true);
    }

    public function news(array $urls): array
    {
        return $this->invoke('bytelyon-article-extractor', compact('urls'));
    }

    public function page(string $url, bool $includeLinks = true): array
    {
        return $this->invoke('bytelyon-page-scraper', compact('url', 'includeLinks'));
    }

    public function serp(string $query, Proxy $proxy): array
    {
        return $this->invoke('bytelyon-serp-scraper', [
            'query' => $query,
            'proxy' => $proxy->toPayload(),
            'geoip' => true,
        ]);
    }
}
