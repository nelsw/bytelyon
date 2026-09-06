<?php

namespace App\Services\Bots\News\Rss;

use App\Enums\NewsSource;
use Carbon\CarbonInterface;
use Carbon\Exceptions\InvalidDateException;
use Illuminate\Container\Attributes\Singleton;
use Illuminate\Http\Client\ConnectionException;
use Illuminate\Http\Client\RequestException;
use Illuminate\Support\Carbon;
use Illuminate\Support\Facades\Http;

#[Singleton]
readonly class BingRssService
{
    private const string RSS_URL = "https://www.bing.com/news/search";

    /**
     * @throws RequestException
     * @throws ConnectionException
     */
    public function fetch(string $query, CarbonInterface $last, array $blacklist): array {

        $body = Http::get(url: self::RSS_URL, query: [
            'q' => $query,
            'format' => 'rss',
        ])->throw()->body();

        $arr = [];
        $xml = simplexml_load_string($body);
        for ($i = 0; $i < count($xml->channel->item); $i++) {
            $item = $xml->channel->item[$i];

            try {
                $date = Carbon::parse((string)$item->pubDate);
            } catch (InvalidDateException) {
                continue;
            }
            if ($date->isBefore($last)) {
                continue;
            }

            $title = (string)$item->title;
            $description = (string)$item->description;
            foreach ($blacklist as $keyword) {
                if (str_contains($title, $keyword) || str_contains($description, $keyword)) {
                    continue 2;
                }
            }

            $arr[] = [
                'title' => $title,
                'description' => $description,
                'pubDate' => $date,
                'publisher' => NewsSource::BingNews->value,
                'source' => (string)$item->xpath('//News:Source')[0],
                'img_url' => (string)$item->xpath('//News:Image')[0],
                'url' => $this->decode((string)$item->link),
            ];
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
