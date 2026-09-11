<?php

namespace App\Services;

use App\Enums\NewsSource;
use App\Models\Bot;
use Carbon\Exceptions\InvalidDateException;
use Illuminate\Container\Attributes\Singleton;
use Illuminate\Support\Carbon;

#[Singleton]
readonly class BingRssService
{
    public function __construct(private XmlService $xmlService) {}

    public function fetch(Bot $bot): array
    {

        $xml = $this->xmlService->fetch('https://www.bing.com/news/search', [
            'q' => $bot->query,
            'format' => 'rss',
        ]);

        $arr = [];
        foreach ($xml->channel->item as $item) {

            try {
                $date = Carbon::parse((string) $item->pubDate);
            } catch (InvalidDateException $e) {
                continue;
            }
            if ($date->isBefore($bot->lastRunAt())) {
                continue;
            }

            foreach ($bot->blacklist() as $keyword) {
                if (str_contains((string) $item->title, $keyword) ||
                    str_contains((string) $item->description, $keyword)) {
                    continue 2;
                }
            }

            $arr[$this->decode((string) $item->link)] = [
                'title' => (string) $item->title,
                'description' => (string) $item->description,
                'published_at' => (string) $item->pubDate,
                'publisher' => NewsSource::BingNews->value,
                'source' => (string) $item->xpath('(//News:Source)[0]'),
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
