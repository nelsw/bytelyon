<?php

namespace App\Services;

use App\Support\Rss\BingRssItem;
use App\Support\Rss\GoogleRssItem;
use Illuminate\Container\Attributes\Singleton;
use Illuminate\Support\Facades\Http;

#[Singleton]
class RssService
{
    public function news(string $query): array
    {
        return array_merge(
            $this->bing($query),
            $this->google($query),
        );
    }

    protected function bing(string $query): array
    {
        return $this->items(BingRssItem::class, 'https://www.bing.com/news/search', [
            'q' => $query,
            'format' => 'rss',
        ]);
    }

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

        return rescue(function() use ($class, $url, $query) {
            return (array) simplexml_load_string(
                data: Http::get($url, $query)->throw()->body(),
                class_name: $class,
            )->xpath('//item');
        }, []);
    }
}
