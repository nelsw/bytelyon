<?php

namespace App\Services;

use Illuminate\Container\Attributes\Singleton;
use Illuminate\Process\Exceptions\ProcessTimedOutException;
use Illuminate\Support\Facades\Log;
use Illuminate\Support\Facades\Process;
use RuntimeException;

#[Singleton]
readonly class ProcessService
{
    public function run(array $args): bool
    {
        $cmd = implode(' ', $args);

        Log::debug('ProcessService::run', ['cmd' => $cmd]);

        try {
            $res = Process::run($args, function (string $type, string $output) use ($cmd) {
                Log::debug('ProcessService::run', [
                    'cmd' => $cmd,
                    'type' => $type,
                    'out' => $output,
                ]);
            });
        } catch (ProcessTimedOutException|RuntimeException $ex) {
            Log::error("ProcessService::run - {$ex->getMessage()}", compact('cmd', 'ex'));
            return false;
        }

        Log::debug('ProcessService::run', [
            'cmd' => $cmd,
            'err' => $res->errorOutput(),
            'ok' => $res->successful(),
            'out' => $res->output(),
        ]);

        return $res->successful();
    }
}
