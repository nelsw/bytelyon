<?php

namespace App\Facades;

use App\Data\Lambda\Payload;
use App\Services\LambdaService;
use Illuminate\Support\Facades\Facade;

/**
 * @method static array<string, mixed> invoke(string $functionName, array<string, mixed> $input = [])
 * @method static Payload scrape(string $url)
 */
class Lambda extends Facade
{
    protected static function getFacadeAccessor(): string
    {
        return LambdaService::class;
    }
}
