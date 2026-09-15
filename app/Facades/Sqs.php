<?php

namespace App\Facades;

use App\Services\SqsService;
use Illuminate\Support\Facades\Facade;

/**
 * @method static void enqueueScrape(string $type, int $id, array $fields = [])
 */
class Sqs extends Facade
{
    protected static function getFacadeAccessor(): string
    {
        return SqsService::class;
    }
}
