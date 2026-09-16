<?php

namespace App\Actions\Model;

use App\Models\Bot;
use Log;

class UpdateBotRunTimestamp
{
    public function __invoke(Bot $val, ?int $key = null): void
    {
        $val->update(['last_run_at' => now()->utc()]);
        Log::debug('Bot last_run_at updated');
    }
}
