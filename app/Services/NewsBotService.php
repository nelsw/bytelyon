<?php

namespace App\Services;

use App\Contracts\RssItem;
use App\Data\Rss\BingRssItem;
use App\Data\Rss\GoogleRssItem;
use App\Events\BotResultsPersisted;
use App\Models\Bot;
use Illuminate\Container\Attributes\Singleton;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Support\Collection;
use Illuminate\Support\Facades\Log;
use Throwable;

#[Singleton]
readonly class NewsBotService extends RssService
{
    public function __construct(
        private LambdaService $lambdaService,
    ) {}

    public function run(Bot $bot): void
    {
        $items = collect()
            ->merge($this->bing($bot->query))
            ->merge($this->google($bot->query))
            ->filter(fn (RssItem $item) => $bot->lastRunAt()->isBefore($item->publishedAt()))
            ->reject(fn (RssItem $item) => $bot->blacklisted($item->title(), $item->description()))
            ->transform(fn (RssItem $item) => [$item->url() => $item->toArray()])
            ->collapse();

        $pages = $items
            ->chunk(5)
            ->transform(fn (Collection $chunk) => $chunk->pluck('url')->all())
            ->map(fn (array $urls) => $this->lambdaService->news($urls))
            ->collapse();

        $saved = $pages->map(function (array|string $page) use ($bot, $items): ?Model {
            try {
                if (is_string($page)) {
                    Log::warning("NewsBotService#run - $page");
                    return null;
                }

                return $bot->articles()->updateOrCreate(
                    attributes: ['url' => $page['url']],
                    values: [
                        ...$items->get($page['url']),
                        ...$page,
                    ]
                );
            } catch (Throwable $e) {
                Log::warning("NewsBotService#run {$e->getMessage()}", ['page' => $page]);
                return null;
            }
        })->filter();

        Log::info('NewsBotService::run', [
            'query' => $bot->query,
            'items' => $items->count(),
            'pages' => $pages->count(),
            'saved' => $saved->count(),
        ]);

        if ($saved->isNotEmpty()) {
            BotResultsPersisted::dispatch($bot, __(':count new article(s) found for ":query".', [
                'count' => $saved->count(),
                'query' => $bot->query,
            ]));
        }
    }

    private function bing(string $query): array
    {
        return $this->items(
            BingRssItem::class,
            'https://www.bing.com/news/search',
            [
                'q' => $query,
                'format' => 'rss',
            ]);
    }

    private function google(string $query): array
    {
        return $this->items(
            GoogleRssItem::class,
            'https://news.google.com/rss/search',
            [
                'q' => $query,
                'hl' => 'en-US',
                'gl' => 'US',
                'ceid' => 'US:en',
            ]);
    }
}
