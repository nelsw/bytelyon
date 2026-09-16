<?php

namespace App\Services;

use App\Support\Rss\BingRssItem;
use App\Support\Rss\GoogleRssItem;
use App\Support\Rss\RssItem;
use Illuminate\Container\Attributes\Singleton;
use Illuminate\Support\Facades\Http;

#[Singleton]
class RssService
{
    /** @return RssItem[] */
    public function news(string $query): array
    {
        return array_merge(
            $this->bing($query),
            $this->google($query),
        );
    }

    /** @return BingRssItem[] */
    protected function bing(string $query): array
    {
        return $this->items(BingRssItem::class, 'https://www.bing.com/news/search', [
            'q' => $query,
            'format' => 'rss',
        ]);
    }

    /** @return GoogleRssItem[] */
    protected function google(string $query): array
    {
        return $this->items(GoogleRssItem::class, 'https://news.google.com/rss/search', [
            'q' => $query,
            'hl' => 'en-US',
            'gl' => 'US',
            'ceid' => 'US:en',
        ]);
    }

    private function items(string $class, string $url, array $query): array
    {
        return (array) simplexml_load_string(
            data: rescue(fn () => Http::get($url, $query)->throw()->body(), []),
            class_name: $class,
        )->xpath('//item');
    }
}
