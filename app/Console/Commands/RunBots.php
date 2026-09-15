<?php

namespace App\Console\Commands;

use App\Enums\BotType;
use App\Jobs\NewsBotJob;
use App\Jobs\SearchBotJob;
use App\Jobs\SitemapBotJob;
use App\Models\Bot;
use Illuminate\Console\Attributes\Description;
use Illuminate\Console\Attributes\Signature;
use Illuminate\Console\Command;

#[Signature('run:bots')]
#[Description('Dispatch runnable (enabled & ready) bots.')]
class RunBots extends Command
{
    public function handle(): void
    {
        $bots = Bot::query()->enabled()->ready();
        if ($bots->doesntExist()) {
            $this->info('no runnable bots found');
            return;
        }
        $bots->each(fn (Bot $bot) => match ($bot->type) {
            BotType::News => NewsBotJob::dispatch($bot),
            BotType::Search => SearchBotJob::dispatch($bot),
            BotType::Sitemap => SitemapBotJob::dispatch($bot),
        });
        $this->info("bots dispatched : {$bots->count()}");
    }
}
