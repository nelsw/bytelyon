<?php

namespace App\Console\Commands;

use App\Jobs\BotJob;
use App\Models\Bot;
use Illuminate\Console\Attributes\Description;
use Illuminate\Console\Attributes\Signature;
use Illuminate\Console\Command;
use Illuminate\Database\Eloquent\ModelNotFoundException;

#[Signature('work:bot {id}')]
#[Description('Dispatch bot job by id.')]
class WorkBot extends Command
{
    public function handle(): void
    {
        try {
            $bot = Bot::query()->findOrFail($this->argument('id'));
        } catch (ModelNotFoundException) {
            $this->warn('Bot not found');
            return;
        }

        BotJob::dispatch($bot);

        $this->info('Bot job dispatched');
    }
}
