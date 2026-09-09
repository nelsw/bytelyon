<?php

namespace App\Services;

use App\Models\Bot;
use Illuminate\Container\Attributes\Singleton;

#[Singleton]
readonly class BotProcessService
{
    public function __construct(private ProcessService $service){}

    public function query(Bot $bot, string $query): bool
    {
        return $this->service->run([
            'uv', 'run', base_path("scripts/{$bot->type->value}.py"),
            '-p', storage_path("app/private/{$bot->type->value}/$bot->id"),
            '-q', $query,
            '--headless'
        ]);
    }

    public function url(Bot $bot, string $url): bool
    {
        return $this->service->run([
            'uv', 'run', base_path("scripts/{$bot->type->value}.py"),
            '-p', storage_path("app/private/{$bot->type->value}/$bot->id"),
            '-u', $url,
            '--headless'
        ]);
    }
}

