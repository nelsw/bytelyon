<?php

namespace App\Traits\Services;

use Aws\Sqs\SqsClient;

trait HasSqsClient
{
    protected function initializeHasSqsClient(): void
    {
        $this->client = new SqsClient(config('services.sqs'));
        $this->queueUrl = config('services.sqs.scrape_jobs_queue_url');
    }
}
