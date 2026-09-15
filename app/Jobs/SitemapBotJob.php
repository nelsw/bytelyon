<?php

namespace App\Jobs;

use App\Facades\Sqs;
use Illuminate\Support\Facades\Log;

class SitemapBotJob extends BaseBotJob
{
    private const int CRAWL_DEPTH = 5;

    /**
     * Enqueues a single scrape job for the sitemap's root URL onto the
     * shared bytelyon-scrape-jobs SQS queue — the crawl continues itself
     * from there. See ScrapeJobController::sitemap(), which enqueues one
     * further job per newly-discovered link (bounded by `depth`,
     * decrementing each hop) every time a page finishes scraping. This is
     * the async equivalent of the old synchronous recursive crawl()/scrape()
     * pair (see git history), with the persisted `pages` table standing in
     * for what used to be an in-memory $done set scoped to a single job
     * execution.
     *
     * Known behavior change: only the root URL is unconditionally
     * re-visited on every bot run now. A previously-discovered page's
     * content (title/meta/screenshot) only refreshes if it's re-reachable
     * via some newly-discovered path this run — it's no longer proactively
     * re-crawled just because it was seen on a prior run. Acceptable
     * trade-off for now; revisit if staleness becomes a real problem.
     */
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
