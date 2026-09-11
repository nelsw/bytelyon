<?php

namespace App\Data\Rss;

use App\Enums\NewsSource;
use Illuminate\Support\Arr;

class BingRssItem extends BaseRssItem
{
    public function publisher(): string
    {
        return (string) Arr::first(array: $this->xpath('//News:Source'), default: '');
    }

    public function source(): string
    {
        return NewsSource::BingNews->value;
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

    function imageSrc(): string
    {
        return (string) Arr::first(array: $this->xpath('//News:Source'), default: '');
    }
}
