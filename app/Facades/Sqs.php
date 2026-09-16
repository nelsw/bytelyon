<?php

namespace App\Facades;

use App\Services\SqsService;
use Closure;
use Illuminate\Support\Facades\Facade;

/**
 * @method static void enqueue(array<string, mixed> $fields = [])
 * @method static Closure enqueueFN(array<string, mixed> $fields = [])
 */
class Sqs extends Facade
{
    protected static function getFacadeAccessor(): string
    {
        return SqsService::class;
    }
}
