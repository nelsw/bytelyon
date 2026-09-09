<?php

namespace App\Services\Bots\News\Rss;

use App\Enums\NewsSource;
use App\Models\Article;
use App\Models\Bot;
use App\Services\XmlService;
use Carbon\Exceptions\InvalidDateException;
use Illuminate\Container\Attributes\Singleton;
use Illuminate\Support\Carbon;
use Illuminate\Support\Facades\Log;

#[Singleton]
readonly class BingRssService
{
    public function __construct(private XmlService $xmlService){}


    public function fetch(Bot $bot): array {

        $xml = $this->xmlService->fetch("https://www.bing.com/news/search", [
            'q' => $bot->query,
            'format' => 'rss',
        ]);

        $arr = [];
        foreach ($xml->channel->item as $item) {

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

            $arr[] = [
                'title' => (string)$item->title,
                'description' => (string)$item->description,
                'published_at' => (string)$item->pubDate,
                'publisher' => NewsSource::BingNews->value,
                'source' => $source,
                'img_url' => $image,
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
