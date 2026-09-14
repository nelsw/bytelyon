<?php

namespace App\Services;

use App\Contracts\RssItem;
use App\Data\Html\Page;
use App\Events\BotResultsPersisted;
use App\Facades\Lambda;
use App\Models\Bot;
use Illuminate\Container\Attributes\Singleton;
use Illuminate\Support\Facades\Log;
use Illuminate\Support\Facades\Storage;
use Throwable;

#[Singleton]
readonly class NewsBotService extends RssService
{
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

            $payload = Lambda::scrape($item->url());
            $page = new Page($item->url(), Storage::disk('s3')->get($payload->contentKey));

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
}
