<?php

namespace App\Services;

use App;
use Aws\Sqs\SqsClient;
use Illuminate\Container\Attributes\Singleton;

#[Singleton]
class SqsService
{
    protected SqsClient $client;

    protected string $queueUrl = '';

    public function __construct()
    {
        $this->client = new SqsClient(config('services.sqs'));
        $this->queueUrl = config('services.sqs.scrape_jobs_queue_url');
    }

    public function enqueueScrape(string $type, int $id, array $fields = []): void
    {
        if (App::runningUnitTests() || !App::isLocal()) {
            return;
        }
        $this->client->sendMessage([
            'QueueUrl' => $this->queueUrl,
            'MessageBody' => json_encode([
                'type' => $type,
                'id' => $id,
                ...$fields,
            ]),
        ]);
    }
}
