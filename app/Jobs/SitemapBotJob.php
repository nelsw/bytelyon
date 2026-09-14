<?php

namespace App\Jobs;

use App\Actions\UpdateOrCreateSitemapPage;
use App\Actions\UpdateSitemap;
use App\Events\BotResultsPersisted;
use App\Facades\Lambda;
use Illuminate\Support\Facades\Log;

class SitemapBotJob extends BaseBotJob
{
    public function handle(): void {

        /** @var array<string, bool> $urls */
        $urls = $this->crawl( 5, [], ["https://{$this->bot->query}" => false]);
        (new UpdateSitemap)($this->bot, $urls);

        $this->bot->update(['last_run_at' => now()->utc()]);
        Log::info('BotJob', [
            'id' => $this->bot->id,
            'type' => $this->bot->type->value,
            'query' => $this->bot->query,
            'urls' => count($urls),
        ]);

        if (count($urls) > 0) {
            BotResultsPersisted::dispatch($this->bot, __(':count page(s) crawled for ":domain".', [
                'count' => count($urls),
                'domain' => $this->bot->query,
            ]));
        }
    }

    /** @return  array<string, bool>  */
    public function scrape(string $url): array
    {
        Log::debug("scraping $url");
        $payload = Lambda::scrape($url);
        (new UpdateOrCreateSitemapPage)($this->bot, $url, $payload->html->title(), $payload->screenshotKey, $payload->html->meta());
        $links = $payload->html->links();
        $count = count($links);
        Log::debug("scraped $url ($count)");
        return $links;
    }

    /**
     * @param array<string, bool> $done
     * @param array<string, bool> $todo
     * @return array<string, bool>
     *
     */
    public function crawl(int $depth, array $done, array $todo): array
    {
        Log::debug('crawling', [
            'depth' => $depth,
            'done' => count($done),
            'todo' => count($todo),
        ]);

        if ($depth < 0 || empty($todo)) {
            return $done;
        }

        $next = [];
        foreach ($todo as $key => $value) {
            if (isset($done[$key])) {
                continue;
            }
            $next = [
                ...$next,
                ...$this->scrape($key),
            ];
            $done[$key] = true;
        }
        Log::debug('crawled', [
            'depth' => $depth,
            'done' => count($done),
            'next' => count($next),
        ]);
        return $this->crawl($depth - 1, $done, $next);
    }
}
