<?php

namespace App\Console\Commands;

use App\Enums\BotType;
use App\Models\Bot;
use App\Models\Sitemap;
use App\Services\Bots\News\NewsBotService;
use App\Services\Bots\Search\SearchBotService;
use App\Services\Bots\Sitemap\SitemapService;
use Illuminate\Console\Attributes\Signature;
use Illuminate\Console\Command;

#[Signature('bot:sitemap {domain}')]
class RunSitemapBot extends Command
{

    public function handle(SitemapService $service): void
    {
        $sitemap = Sitemap::factory()->createQuietly([
            'domain' => $this->argument('domain'),
        ]);

        dump($sitemap->toPrettyJson());

        $ok = $service->run($sitemap);
    }
}
