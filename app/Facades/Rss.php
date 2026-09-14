<?php

namespace App\Facades;

use App\Data\Rss\BaseRssItem;
use App\Services\RssService;
use Illuminate\Support\Facades\Facade;

/**
 * @method static BaseRssItem[] news(string $query)
 */
class Rss extends Facade
{
    protected static function getFacadeAccessor(): string
    {
        return RssService::class;
    }
}
