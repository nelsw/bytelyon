<?php

namespace App\Jobs;

use App\Actions\Model\UpdateBotRunTimestamp;
use App\Enums\BotType;
use App\Events\BotResultsPersisted;
use App\Facades\Rss;
use App\Facades\Sqs;
use App\Models\Article;
use App\Models\Bot;
use App\Support\Rss\RssItem;
use Illuminate\Bus\Batchable;
use App\Models\Proxy;
use App\Support\Rss\BaseRssItem;
use Illuminate\Contracts\Queue\ShouldBeUnique;
use Illuminate\Contracts\Queue\ShouldQueue;
use Illuminate\Foundation\Queue\Queueable;
use Illuminate\Queue\SerializesModels;
use Illuminate\Support\Collection;
use Illuminate\Support\Facades\Context;
use Illuminate\Support\Facades\Log;
use Throwable;

class BotJob implements ShouldBeUnique, ShouldQueue
{
    use Batchable, Queueable, SerializesModels;

    public function __construct(
        public readonly Bot $bot,
    ) {}

    public function uniqueId(): string
    {
        return Context::id();
    }

    public function prepareForDispatch(): bool
    {
        return $this->bot->isRunnable();
    }

    public function handle(): void
    {
        Log::debug('BotJob');

        switch ($this->bot->type) {
            case BotType::Search:
                Sqs::enqueue(Context::all());
                break;
            case BotType::Sitemap:
                Sqs::enqueue([...Context::all(), ...['url' => "https://{$this->bot->query}", 'depth' => 5]]);
                break;
            case BotType::News:
                collect(Rss::news($this->bot->query))
                    ->filter(fn (RssItem $item) => $this->bot->lastRunAt()->isBefore($item->publishedAt()))
                    ->reject(fn (RssItem $item) => $this->bot->blacklisted($item->title(), $item->description()))
                    ->transform(fn (RssItem $item) => $this->bot->articles()->updateOrCreate(
                        attributes: ['url' => $item->url()],
                        values: $item->toArray(),
                    ))->each(fn (Article $article) => Sqs::enqueue([
                        'type' => BotType::News->value,
                        'id' => $this->bot->id,
                        'url' => $article->url,
                    ]))->whenNotEmpty(function (Collection $collection) {
                        BotResultsPersisted::dispatch($this->bot, __(':count new article(s) queued for ":query".', [
                            'count' => $collection->count(),
                            'query' => $this->bot->query,
                        ]));
                    });
        }
        (new UpdateBotRunTimestamp)($this->bot);
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

    /**
     * @return array<string, array<string, mixed>>
     */
    protected function proxyFields(): array
    {
        $proxy = $this->bot->randomProxy();

        if ($proxy === null) {
            return [];
        }

        return ['proxy' => $this->proxyPayload($proxy)];
    }

    /** @return array<string, mixed> */
    protected function proxyPayload(Proxy $proxy): array
    {
        return [
            'scheme' => $proxy->scheme,
            'host' => $proxy->host,
            'port' => $proxy->port,
            'username' => $proxy->username,
            'pass' => $proxy->pass,
            'bypass' => $proxy->bypass,
        ];
    }

    public function failed(?Throwable $e): void
    {
        Log::error('BotJob', ['exception' => $e]);
        (new UpdateBotRunTimestamp)($this->bot);
    }
}
