<?php

namespace App\Services\Bots\News\Rss;

use App\Enums\NewsSource;
use App\Models\Article;
use App\Models\Bot;
use App\Models\Page;
use Carbon\CarbonInterface;
use Carbon\Exceptions\InvalidDateException;
use Illuminate\Container\Attributes\Singleton;
use Illuminate\Http\Client\ConnectionException;
use Illuminate\Http\Client\RequestException;
use Illuminate\Support\Carbon;
use Illuminate\Support\Facades\Http;
use Illuminate\Support\Facades\Log;

#[Singleton]
readonly class BingRssService
{
    private const string RSS_URL = "https://www.bing.com/news/search";

    /**
     * @return array<string, Article>
     * @throws RequestException
     * @throws ConnectionException
     */
    public function fetch(Bot $bot): array {

        $body = Http::get(url: self::RSS_URL, query: [
            'q' => $bot->query,
            'format' => 'rss',
        ])->throw()->body();

        $arr = [];
        $xml = simplexml_load_string($body);
        for ($i = 0; $i < count($xml->channel->item); $i++) {
            $item = $xml->channel->item[$i];

            try {
                $date = Carbon::parse((string)$item->pubDate);
            } catch (InvalidDateException $e) {
                Log::warning('BingRssService::fetch', [
                    'exception' => $e,
                    'item' => $item,
                ]);
                continue;
            }
            if ($date->isBefore($bot->last_run_at)) {
                continue;
            }

            foreach ($bot->blacklist() as $keyword) {
                if (str_contains((string)$item->title, $keyword) ||
                    str_contains((string)$item->description, $keyword)) {
                    continue 2;
                }
            }

            $source = '';
            if (sizeof($item->xpath('//News:Source')) > 0) {
                $source = (string)$item->xpath('//News:Source')[0];
            }

            $image = '';
            if (sizeof($item->xpath('//News:Image'))) {
                $image = (string) $item->xpath('//News:Image')[0];
            }

            $arr[] = new Article([
                'title' => (string)$item->title,
                'description' => (string)$item->description,
                'published_at' => $date->toDateTimeString(),
                'publisher' => NewsSource::BingNews->value,
                'source' => $source,
                'img_url' => $image,
                'url' => $this->decode((string)$item->link),
            ]);
        }

        return $arr;
    }

    private function decode(string $link): string
    {
        foreach (explode('&', parse_url($link)['query']) as $param) {
            $keyValue = explode('=', $param);
            if ($keyValue[0] === 'url') {
                return rtrim(urldecode($keyValue[1]), '/');
            }
        }
        return $link;
    }
}
