<?php

namespace App\Facades;

use App\Models\Bot;
use App\Services\GoService;
use Closure;
use Illuminate\Support\Facades\Facade;

/**
 * @method static void pub(Bot|int $bot)
 * @method static void sub(Closure<string> $message)
 */
class Go extends Facade
{
    protected static function getFacadeAccessor(): string
    {
        return GoService::class;
    }
}
