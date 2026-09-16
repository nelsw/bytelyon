<?php

namespace App\Console\Commands;

use App\Jobs\BotJob;
use App\Models\Bot;
use Bus;
use Illuminate\Console\Attributes\Description;
use Illuminate\Console\Attributes\Signature;
use Illuminate\Console\Command;
use Throwable;

#[Signature('run:bots')]
#[Description('Dispatch runnable (enabled & ready) bots.')]
class RunBots extends Command
{
    /**
     * @throws Throwable
     */
    public function handle(): void
    {
        $jobs = Bot::query()
            ->enabled()
            ->ready()
            ->get()
            ->map(fn (Bot $bot) => new BotJob($bot))
            ->all();

        $count = count($jobs);
        $this->info("jobs to run [$count]");
        if ($count > 0) {
            Bus::batch($jobs)->name('run-bots')->dispatch();
        }
    }
}
