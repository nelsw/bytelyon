<?php

namespace App\Services;

use AllowDynamicProperties;
use App;
use App\Traits\Services\HasSqsClient;
use Aws\Sqs\SqsClient;
use Closure;
use Illuminate\Container\Attributes\Singleton;

#[AllowDynamicProperties]
#[Singleton]
class SqsService
{
    use HasSqsClient;

    protected SqsClient $client;

    protected string $queueUrl;

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
