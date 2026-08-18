<?php

namespace App\Console\Commands;

use App\Jobs\BotJob;
use App\Models\Bot;
use Illuminate\Console\Attributes\Description;
use Illuminate\Console\Attributes\Signature;
use Illuminate\Console\Command;
use Illuminate\Support\Collection;

#[Signature('work:bots {type}')]
#[Description('Dispatch bot jobs by type.')]
class WorkBots extends Command
{
    public function handle(): void
    {
        Bot::query()
            ->type($this->argument('type'))
            ->enabled()
            ->ready()
            ->get()
            ->tap(fn (Collection $bots) => $this->info("runnable bots found: {$bots->count()}"))
            ->each(fn (Bot $bot) => BotJob::dispatch($bot));
    }
}
