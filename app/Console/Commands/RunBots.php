<?php

namespace App\Console\Commands;

use App\Jobs\BotJob;
use App\Models\Bot;
use Bus;
use Illuminate\Console\Attributes\Description;
use Illuminate\Console\Attributes\Signature;
use Illuminate\Console\Command;

#[Signature('run:bots')]
#[Description('Dispatch runnable (enabled & ready) bots.')]
class RunBots extends Command
{
    public function handle(): void
    {
        $jobs = Bot::query()
            ->enabled()
            ->ready()
            ->get()
            ->map(fn (Bot $bot) => new BotJob($bot));

        if ($jobs->isEmpty()) {
            $this->info('nothing to run');
            return;
        }

        $batchId = Bus::batch($jobs)
            ->then(fn () => $this->info('bots dispatched'))
            ->catch(fn () => $this->error('bots failed'))
            ->finally(fn () => $this->info('bots finished'))
            ->name('run-bots')
            ->dispatch()
            ->id;

        $batch = Bus::findBatch($batchId);
        do {
            $progress = $batch->progress();
            $this->info("%$progress complete");
        } while ($progress < 100);
    }
}
