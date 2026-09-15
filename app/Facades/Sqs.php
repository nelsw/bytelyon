<?php

namespace App\Facades;

use App\Enums\BotType;
use App\Services\SqsService;
use Illuminate\Support\Facades\Facade;

/**
 * @method static void enqueueScrape(string $type, int $id, array<string, mixed> $fields = [])
 * @method static void enqueue(BotType $type, int $id, array<string, mixed> $fields = [])
 */
class Sqs extends Facade
{
    protected static function getFacadeAccessor(): string
    {
        return SqsService::class;
    }
}
