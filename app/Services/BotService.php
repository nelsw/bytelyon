<?php

namespace App\Services;

use App\Enums\BotType;
use App\Models\Bot;
use Illuminate\Container\Attributes\Singleton;

#[Singleton]
readonly class BotService
{
    public function __construct(
        private NewsBotService $newsBotService,
        private SearchBotService $searchBotService,
        private SitemapBotService $sitemapBotService,
    ) {}

    public function run(Bot $bot): void
    {
        switch ($bot->type) {
            case BotType::News:
                $this->newsBotService->run($bot);
                break;
            case BotType::Search:
                $this->searchBotService->run($bot);
                break;
            case BotType::Sitemap:
                $this->sitemapBotService->run($bot);
                break;
        }
    }
}
