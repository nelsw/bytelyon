<?php

namespace App\Observers;

use App\Enums\BotType;
use App\Jobs\BotJob;
use App\Models\Article;
use App\Models\Bot;

class BotObserver
{
    public function created(Bot $bot): void
    {
        BotJob::dispatch($bot);
    }

    public function updated(Bot $bot): void
    {
        BotJob::dispatch($bot);
    }

    public function deleting(Bot $bot): void
    {
        switch ($bot->type) {
            case BotType::News:
                $bot->articles->each(fn (Article $article) => $article->delete());
                break;
            case BotType::Search:
                $bot->serp?->delete();
                break;
            case BotType::Sitemap:
                $bot->sitemap?->delete();
        }
    }
}
