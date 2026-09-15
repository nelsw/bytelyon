<?php

namespace App\Jobs;

use App\Contracts\RssItem;
use App\Data\Html\Page;
use App\Events\BotResultsPersisted;
use App\Facades\Lambda;
use App\Facades\Rss;
use Illuminate\Support\Facades\Log;
use Illuminate\Support\Facades\Storage;

class NewsBotJob extends BaseBotJob
{
    public function handle(): void
    {

        $items = collect(Rss::news($this->bot->query))
            ->filter(fn (RssItem $item) => $this->bot->lastRunAt()->isBefore($item->publishedAt()))
            ->reject(fn (RssItem $item) => $this->bot->blacklisted($item->title(), $item->description()));

        $pages = 0;
        foreach ($items as $item) {

            $payload = Lambda::scrape($item->url());
            $page = new Page($item->url(), Storage::disk('s3')->get($payload->contentKey));

            $article = [...$item->toArray(), ...$page->toArray()];

            $this->bot->articles()->updateOrCreate(['url' => $item->url()], $article);
            $pages++;
        }

        $this->bot->update(['last_run_at' => now()->utc()]);
        Log::info('BotJob', [
            'id' => $this->bot->id,
            'type' => $this->bot->type->value,
            'query' => $this->bot->query,
            'articles' => $pages,
        ]);

        if ($pages > 0) {
            BotResultsPersisted::dispatch($this->bot, __(':count new article(s) found for ":query".', [
                'count' => $pages,
                'query' => $this->bot->query,
            ]));
        }
    }
}
