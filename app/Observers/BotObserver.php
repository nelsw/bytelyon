<?php

namespace App\Observers;

use App\Enums\BotType;
use App\Jobs\NewsBotJob;
use App\Jobs\SearchBotJob;
use App\Jobs\SitemapBotJob;
use App\Models\Article;
use App\Models\Bot;

class BotObserver
{
    public function created(Bot $bot): void
    {
        switch ($bot->type) {
            case BotType::Search:
                $bot->serp()->create(['query' => $bot->query]);
                SearchBotJob::dispatch($bot);
                return;
            case BotType::Sitemap:
                $bot->sitemap()->create(['domain' => $bot->query]);
                SitemapBotJob::dispatch($bot);
                return;
            case BotType::News:
                NewsBotJob::dispatch($bot);
                return;
        }
    }

    public function updated(Bot $bot): void
    {
        match ($bot->type) {
            BotType::News => SearchBotJob::dispatch($bot),
            BotType::Search => SitemapBotJob::dispatch($bot),
            BotType::Sitemap => NewsBotJob::dispatch($bot),
        };
    }

    public function deleting(Bot $bot): void
    {
        match ($bot->type) {
            BotType::News => $bot->articles->each(fn (Article $article) => $article->delete()),
            BotType::Search => $bot->serp?->delete(),
            BotType::Sitemap => $bot->sitemap?->delete(),
        };
    }
}
