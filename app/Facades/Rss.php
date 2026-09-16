<?php

namespace App\Facades;

use App\Services\RssService;
use App\Support\Rss\RssItem;
use Illuminate\Support\Facades\Facade;

/**
 * @method static RssItem[] news(string $query)
 */
class Rss extends Facade
{
    protected static function getFacadeAccessor(): string
    {
        return RssService::class;
    }
}
