<?php

namespace App\Services;

use App\Contracts\RssItem;
use App\Data\Html\Page;
use App\Data\Rss\BingRssItem;
use App\Data\Rss\GoogleRssItem;
use App\Events\BotResultsPersisted;
use App\Models\Bot;
use Illuminate\Container\Attributes\Singleton;
use Illuminate\Support\Facades\Log;
use Illuminate\Support\Facades\Storage;
use Throwable;

#[Singleton]
readonly class NewsBotService extends RssService
{
    public function __construct(
        private LambdaService $service,
    ) {}

    public function run(Bot $bot): void
    {
        $items = collect()
            ->merge($this->bing($bot->query))
            ->merge($this->google($bot->query))
            ->filter(fn (RssItem $item) => $bot->lastRunAt()->isBefore($item->publishedAt()))
            ->reject(fn (RssItem $item) => $bot->blacklisted($item->title(), $item->description()));

        Log::debug('NewsBotService::run', [
            'query' => $bot->query,
            'items' => $items->count(),
        ]);

        $pages = 0;
        foreach ($items as $item) {

            $out = $this->service->grab($item->url());

            $page = new Page($item->url(), Storage::disk('s3')->get($out['content_key']));

            $article = [...$item->toArray(), ...$page->toArray()];

            try {
                $bot->articles()->updateOrCreate(['url' => $item->url()], $article);
                $pages++;
            } catch (Throwable $e) {
                Log::warning("NewsBotService#run: {$e->getMessage()}", compact('article'));
            }
        }

        Log::debug('NewsBotService::run', [
            'query' => $bot->query,
            'items' => $items->count(),
            'pages' => $pages,
        ]);

        if ($pages > 0) {
            BotResultsPersisted::dispatch($bot, __(':count new article(s) found for ":query".', [
                'count' => $pages,
                'query' => $bot->query,
            ]));
        }
    }

    private function bing(string $query): array
    {
        return $this->items(BingRssItem::class, 'https://www.bing.com/news/search', [
            'q' => $query,
            'format' => 'rss',
        ]);
    }

    private function google(string $query): array
    {
        return $this->items(GoogleRssItem::class, 'https://news.google.com/rss/search', [
            'q' => $query,
            'hl' => 'en-US',
            'gl' => 'US',
            'ceid' => 'US:en',
        ]);
    }
}
