<?php

namespace App\Services;

use Aws\Sqs\SqsClient;
use Illuminate\Container\Attributes\Singleton;

#[Singleton]
readonly class SqsService
{
    private SqsClient $client;

    public function __construct()
    {
        $this->client = new SqsClient(config('services.sqs'));
    }

    /**
     * Enqueue a scrape job for a local worker to pick up (see
     * docker/worker/worker.py). Fire-and-forget: the worker reports the
     * result back via one of the POST /api/scrape-jobs/{type}/{id}/complete
     * endpoints, guarded by a `worker`-ability Sanctum token (see
     * routes/api.php).
     *
     * $type    "serp" | "news" | "sitemap" — picks both the worker's
     *          navigation handler and which Laravel callback route/model
     *          `$id` resolves against.
     * $id      meaning is type-specific: a Serp id for "serp", an Article
     *          id for "news", a Bot id for "sitemap" (see
     *          ScrapeJobController for why each type binds differently).
     * $fields  type-specific navigation input, e.g. ['query' => ...] for
     *          serp or ['url' => ..., 'depth' => ...] for news/sitemap.
     */
    public function enqueueScrape(string $type, int $id, array $fields = []): void
    {
        $this->client->sendMessage([
            'QueueUrl' => config('services.sqs.scrape_jobs_queue_url'),
            'MessageBody' => json_encode([
                'type' => $type,
                'id' => $id,
                ...$fields,
            ]),
        ]);
    }
}
