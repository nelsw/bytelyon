<?php

namespace App\Services\Bots;

use App\Enums\BotType;
use Illuminate\Container\Attributes\Singleton;
use Illuminate\Support\Facades\Process;

#[Singleton]
readonly class ScriptsService
{

    public function pages(
        string $name,
        BotType $type,
        int $botId,
        array $urls,
        bool $headless,
    ): bool {
        $args = [
            ...[
                'uv', 'run', base_path("scripts/$name.py"),
                '-p', storage_path('app/private') . "/$type->value/$botId",
                '-u'
            ],
            ...$urls,
        ];
        if ($headless) {
            $args[] = '--headless';
        }
        return Process::run($args)->successful();
    }

    public function data(BotType $type, int $botId, array $urls, bool $headless): bool
    {
        return $this->pages('data', $type, $botId, $urls, $headless);
    }

    public function content(BotType $type, int $botId, array $urls, bool $headless): bool
    {
        return $this->pages('content', $type, $botId, $urls, $headless);
    }

    public function screenshot(BotType $type, int $botId, array $urls, bool $headless): bool
    {
        return $this->pages('screenshot', $type, $botId, $urls, $headless);
    }

    public function search(int $botId, bool $headless, string $query): bool
    {
        $args = [
            'uv', 'run', base_path('scripts/search.py'),
            '-p', implode('/', [storage_path('app/private'), BotType::Search->value, $botId]),
            '-q', $query
        ];
        if ($headless) {
            $args[] = '--headless';
        }
        return Process::run($args)->successful();
    }
}
