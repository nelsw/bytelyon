<?php

namespace App\Jobs;

use App\Facades\Sqs;
use Illuminate\Support\Facades\Log;

class SitemapBotJob extends BaseBotJob
{
    private const int CRAWL_DEPTH = 5;

    public function handle(): void
    {
        $root = "https://{$this->bot->query}";

        Sqs::enqueueScrape('sitemap', $this->bot->id, ['url' => $root, 'depth' => self::CRAWL_DEPTH]);

        $this->bot->update(['last_run_at' => now()->utc()]);

        Log::info('BotJob: enqueued', [
            'id' => $this->bot->id,
            'type' => $this->bot->type->value,
            'query' => $this->bot->query,
            'root' => $root,
        ]);
    }
}
