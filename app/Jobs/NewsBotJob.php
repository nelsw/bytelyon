<?php

namespace App\Jobs;

use App\Contracts\RssItem;
use App\Events\BotResultsPersisted;
use App\Facades\Rss;
use App\Facades\Sqs;
use Illuminate\Support\Facades\Log;

class NewsBotJob extends BaseBotJob
{
    /**
     * Creates an Article stub per new RSS item (title/description/etc. only)
     * and enqueues a scrape job per item onto the shared
     * bytelyon-scrape-jobs SQS queue instead of blocking on a synchronous
     * Lambda::scrape() call per item. A local worker
     * (docker/worker/worker.py) fetches the actual article page and reports
     * back via POST /api/scrape-jobs/news/{article}/complete, which layers
     * the parsed page content (body/description/img/keywords — see
     * App\Data\Html\Page) on top of the RSS-derived stub, same field
     * precedence as the old synchronous flow (scraped page data overrides
     * RSS feed data for any overlapping key).
     */
    public function handle(): void
    {
        $items = collect(Rss::news($this->bot->query))
            ->filter(fn (RssItem $item) => $this->bot->lastRunAt()->isBefore($item->publishedAt()))
            ->reject(fn (RssItem $item) => $this->bot->blacklisted($item->title(), $item->description()));

        foreach ($items as $item) {
            $article = $this->bot->articles()->updateOrCreate(['url' => $item->url()], $item->toArray());
            Sqs::enqueueScrape('news', $article->id, ['url' => $item->url()]);
        }

        $this->bot->update(['last_run_at' => now()->utc()]);
        Log::info('BotJob: enqueued', [
            'id' => $this->bot->id,
            'type' => $this->bot->type->value,
            'query' => $this->bot->query,
            'articles' => $items->count(),
        ]);

        if ($items->count() > 0) {
            BotResultsPersisted::dispatch($this->bot, __(':count new article(s) queued for ":query".', [
                'count' => $items->count(),
                'query' => $this->bot->query,
            ]));
        }
    }
}
