<?php

namespace App\Facades;

use App\Data\Lambda\Payload;
use App\Models\Proxy;
use App\Services\LambdaService;
use Illuminate\Support\Facades\Facade;

/**
 * @method static Payload scrape(string $url)
 * @method static array<string, mixed> serp(string $query, Proxy $proxy)
 */
class Lambda extends Facade
{
    protected static function getFacadeAccessor(): string
    {
        return LambdaService::class;
    }
}
