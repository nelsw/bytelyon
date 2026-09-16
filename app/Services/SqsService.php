<?php

namespace App\Services;

use AllowDynamicProperties;
use App;
use Aws\Sqs\SqsClient;
use Closure;
use Illuminate\Container\Attributes\Singleton;

#[AllowDynamicProperties]
#[Singleton]
class SqsService
{
    protected SqsClient $client;

    protected string $queueUrl;

    public function __construct()
    {
        $this->client = new SqsClient(config('services.sqs'));
        $this->queueUrl = config('services.sqs.scrape_jobs_queue_url');
    }

    /** @param array<string, mixed> $payload */
    public function enqueue(array $payload = []): void
    {
        $this->enqueueFN()($payload);
    }

    public function enqueueFN(): Closure
    {
        if (App::runningUnitTests()) {
            return fn () => [];
        }

        /** @param array<string, mixed> $payload */
        return fn (array $payload = []) => $this->client->sendMessage([
            'QueueUrl' => $this->queueUrl,
            'MessageBody' => json_encode($payload),
        ]);
    }
}
