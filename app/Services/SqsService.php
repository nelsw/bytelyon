<?php

namespace App\Services;

use App;
use Aws\Sqs\SqsClient;
use Closure;
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

    public function enqueueFN(): Closure
    {
        if (App::runningUnitTests()) {
            return fn () => [];
        }

        /** @param array<string, mixed> $payload */
        return fn(array $payload = []) => $this->client->sendMessage([
            'QueueUrl' => $this->queueUrl,
            'MessageBody' => json_encode($payload),
        ]);
    }

    /** @param array<string, mixed> $payload */
    public function enqueue(array $payload = []): void
    {
        $this->enqueueFN()($payload);
    }
}
