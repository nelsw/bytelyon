<?php

namespace App\Services;

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

    public function invoke(string $functionName, array $payload): array
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
        return $this->invoke('bytelyon-article-extractor', ['urls' => $urls]);
    }
}
