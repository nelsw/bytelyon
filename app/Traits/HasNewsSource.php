<?php

namespace App\Traits;

use App\Enums\NewsSource;
use App\Support\Rss\BaseRssItem;
use App\Support\Rss\BingRssItem;
use App\Support\Rss\GoogleRssItem;

trait HasNewsSource
{
    public function source(): string
    {
        return match (static::class) {
            BingRssItem::class => NewsSource::BingNews->value,
            GoogleRssItem::class => NewsSource::GoogleNews->value,
            default => BaseRssItem::class,
        };
    }
}
