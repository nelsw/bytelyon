<?php

namespace App\Facades;

use App\Services\S3Service;
use Illuminate\Support\Facades\Facade;

/**
 * @method static void del(string $key)
 * @method static string url(string $key, float|int $ttl = 900)
 */
class S3 extends Facade
{
    protected static function getFacadeAccessor(): string
    {
        return S3Service::class;
    }
}
