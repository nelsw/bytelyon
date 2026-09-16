<?php

namespace App\Support\Rss;

use Illuminate\Support\Arr;

class BingRssItem extends RssItem
{
    public function publisher(): string
    {
        return (string) Arr::first(array: $this->xpath('//News:Source'), default: '');
    }

    public function url(): string
    {
        foreach (explode('&', parse_url((string) $this->link)['query']) as $param) {
            $keyValue = explode('=', $param);
            if ($keyValue[0] === 'url') {
                return rtrim(urldecode($keyValue[1]), '/');
            }
        }
        return (string) $this->link;
    }

    public function imageSrc(): string
    {
        return (string) Arr::first(array: $this->xpath('//News:Image'), default: '');
    }
}
