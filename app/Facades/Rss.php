<?php

namespace App\Facades;

use App\Services\RssService;
use App\Support\Rss\BaseRssItem;
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
