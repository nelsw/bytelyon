<?php

namespace App\Jobs;

use App\Enums\BotType;
use App\Events\BotResultsPersisted;
use App\Facades\Rss;
use App\Facades\Sqs;
use App\Models\Bot;
use App\Support\Rss\BaseRssItem;
use Illuminate\Contracts\Queue\ShouldBeUnique;
use Illuminate\Contracts\Queue\ShouldQueue;
use Illuminate\Foundation\Queue\Queueable;
use Illuminate\Queue\SerializesModels;
use Illuminate\Support\Facades\Log;
use Throwable;

class BotJob implements ShouldBeUnique, ShouldQueue
{
    use Queueable, SerializesModels;

    public function __construct(
        public readonly Bot $bot,
    ) {
        $this->onQueue('default');
    }

    public function uniqueId(): string
    {
        return strval($this->bot->id);
    }

    public function prepareForDispatch(): bool
    {
        return $this->bot->isRunnable();
    }

    public function handle(): void
    {
        Log::debug('BotJob::handle', $this->bot->toArray());

        switch ($this->bot->type) {
            case BotType::Search:
                $this->bot->serp()->create(['query' => $this->bot->query]);
                $this->handleSearch();
                break;
            case BotType::Sitemap:
                $this->bot->sitemap()->create(['domain' => $this->bot->query]);
                $this->handleSitemap();
                break;
            case BotType::News:
                $this->handleNews();
                break;
        }
    }

    public function handleNews(): void
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

    public function handleSearch(): void
    {
        Sqs::enqueueScrape('serp', $this->bot->serp->id, ['query' => $this->bot->query]);
    }

    public function handleSitemap(): void
    {
        $root = "https://{$this->bot->query}";

        Sqs::enqueueScrape('sitemap', $this->bot->id, ['url' => $root, 'depth' => 5]);
    }

    public function failed(?Throwable $e): void
    {
        $this->bot->update(['last_run_at' => now()->utc()]);
        Log::error('BotJob', [
            'exception' => $e,
            'id' => $this->bot->id,
            'type' => $this->bot->type->value,
            'query' => $this->bot->query,
        ]);
    }
}
