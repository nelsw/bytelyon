<?php

namespace App\Console\Commands;

use App\Enums\BotType;
use App\Models\Bot;
use App\Services\Bots\News\NewsBotService;
use App\Services\Bots\Search\SearchBotService;
use Illuminate\Console\Attributes\Signature;
use Illuminate\Console\Command;

#[Signature('bot:news {query}')]
class RunNewsBot extends Command
{

    public function handle(
        NewsBotService $service,
    ): void
    {
       $bot = Bot::factory()
           ->news()
           ->headless()
           ->createQuietly([
               'blacklist' => '',
               'query' => $this->argument('query'),
               'last_run_at' => now()->subYear(),
           ]);


        dump($bot->toPrettyJson());

        $ok = $service->run($bot);

    }
}
