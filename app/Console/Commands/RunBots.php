<?php

namespace App\Console\Commands;

use App\Jobs\BotJob;
use App\Models\Bot;
use Illuminate\Console\Attributes\Description;
use Illuminate\Console\Attributes\Signature;
use Illuminate\Console\Command;
use Illuminate\Support\Collection;

#[Signature('run:bots')]
#[Description('Dispatch runnable (enabled & ready) bots.')]
class RunBots extends Command
{
    public function handle(): void
    {
        $bots = Bot::query()->enabled()->ready();
        if ($bots->doesntExist()) {
            $this->info("no runnable bots found");
            return;
        }
        $bots->each(fn (Bot $bot) => BotJob::dispatch($bot));
        $this->info("bots dispatched : {$bots->count()}");
    }
}
