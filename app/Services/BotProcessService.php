<?php

namespace App\Services;

use App\Models\Bot;
use Illuminate\Container\Attributes\Singleton;
use Illuminate\Process\Exceptions\ProcessTimedOutException;
use Illuminate\Support\Facades\App;
use Illuminate\Support\Facades\Log;
use Illuminate\Support\Facades\Process;
use RuntimeException;

#[Singleton]
readonly class BotProcessService
{
    private string $uv;
    public function __construct(
        private ProcessService $service,
    ) {
        $this->uv = App::isProduction() ? '/home/forge/.local/bin/uv' : 'uv';
    }

    public function query(Bot $bot, string $query): bool
    {
        return $this->service->run([
            $this->uv, 'run', base_path("scripts/{$bot->type->value}.py"),
            '-p', storage_path("app/private/{$bot->type->value}/$bot->id"),
            '-q', $query,
            '--headless',
        ]);
    }

    public function url(Bot $bot, string $url): bool
    {
        return $this->service->run([
            $this->uv, 'run', base_path("scripts/{$bot->type->value}.py"),
            '-p', storage_path("app/private/{$bot->type->value}/$bot->id"),
            '-u', $url,
            '--headless',
        ]);
    }

    public function urls(Bot $bot, array $urls): bool
    {
        $args = [
            $this->uv, 'run', base_path("scripts/{$bot->type->value}.py"),
            '-p', storage_path("app/private/{$bot->type->value}/$bot->id"),
            '-u',
        ];
        $args = [
            ...$args,
            ...$urls,
        ];
        $args[] = '--headless';
        return $this->service->run($args);
    }

    public function news(array $urls): string|array
    {
        return $this->run([
            ...[$this->uv, 'run', base_path("scripts/news.py")],
            ...$urls,
        ]);
    }

    private function run(array $args): string|array
    {
        try {
            $res = Process::run($args);
        } catch (ProcessTimedOutException|RuntimeException $ex) {
            return $ex->getMessage();
        }

        if ($res->successful()) {
            return json_decode($res->output(), true);
        }
        return $res->errorOutput();
    }
}
