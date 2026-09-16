<?php

namespace App\Jobs;

use App\Events\BotResultsPersisted;
use App\Facades\Rss;
use App\Facades\Sqs;
use App\Support\Rss\BaseRssItem;
use Illuminate\Support\Facades\Log;

class NewsBotJob extends BaseBotJob
{
    public function handle(): void
    {
        $items = collect(Rss::news($this->bot->query))
            ->filter(fn (BaseRssItem $item) => $this->bot->lastRunAt()->isBefore($item->publishedAt()))
            ->reject(fn (BaseRssItem $item) => $this->bot->blacklisted($item->title(), $item->description()));

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
