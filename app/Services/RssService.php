<?php

namespace App\Services;

use App\Support\Rss\BingRssItem;
use App\Support\Rss\GoogleRssItem;
use Illuminate\Http\Client\ConnectionException;
use Illuminate\Http\Client\RequestException;
use Illuminate\Support\Facades\Http;

readonly class RssService
{
    private function items(string $class, string $url, array $query): array
    {
        try {
            $body = Http::get($url, $query)->throw()->body();
        } catch (RequestException|ConnectionException $e) {
            return [];
        }
        return (array) simplexml_load_string($body, $class)->xpath('//item');
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

    public function news(string $query): array
    {
        return array_merge(
            $this->bing($query),
            $this->google($query),
        );
    }
}
