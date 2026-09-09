<?php

namespace App\Traits;

use App\Enums\BotType;
use App\Models\Bot;
use Illuminate\Database\Eloquent\Relations\BelongsTo;
use Illuminate\Process\Exceptions\ProcessTimedOutException;
use Illuminate\Support\Arr;
use Illuminate\Support\Facades\Log;
use Illuminate\Support\Facades\Process;
use RuntimeException;

trait HasBotProcess
{
    use HasBot;

    public function process(array $args): bool
    {
        // only uv .py scripts supported for bots at this point
        $input = [
            'uv', 'run', base_path("scripts/{$this->bot->type->value}.py"),
            '-p', storage_path("app/private/{$this->bot->type->value}/$this->id"),
        ];

        // scripts always require additional arguments, but it may change
        if (count($args) > 0) {
            $input = [...$input, ...$args];
        }

        // headless supported with a single flag (only)
        if ($this->bot->headless) {
            $input[] = '--headless';
        }

        // join the arguments for clear logging
        $cmd = implode(' ', $input);

        Log::debug('HasProcess::run', ['cmd' => $cmd]);
        try {
            $res = Process::run($input, function (string $type, string $output) use ($cmd) {
                Log::debug('HasProcess::run', [
                    'cmd' => $cmd,
                    'type' => $type,
                    'out' => $output,
                ]);
            });
        } catch (ProcessTimedOutException|RuntimeException $ex) {
            Log::error("HasProcess::run - {$ex->getMessage()}", compact('cmd', 'ex'));
            return false;
        }

        Log::debug('SearchBotService::run', [
            'cmd' => $cmd,
            'err' => $res->errorOutput(),
            'ok' => $res->successful(),
            'out' => $res->output(),
        ]);

        return $res->successful();
    }
}
