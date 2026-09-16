<?php

namespace App\Observers;

use App\Enums\BotType;
use App\Jobs\BotJob;
use App\Models\Article;
use App\Models\Bot;
use Illuminate\Support\Facades\Context;

class BotObserver
{
    public function created(Bot $bot): void
    {
        Context::add([
            'id' => $bot->id,
            'type' => $bot->type->value,
            'query' => $bot->query,
        ]);

        match ($bot->type) {
            BotType::Search => $bot->serp()->create(['query' => $bot->query]),
            BotType::Sitemap => $bot->sitemap()->create(['domain' => $bot->query]),
            BotType::News => null,
        };

        BotJob::dispatchAfterResponse($bot);
    }

    public function updated(Bot $bot): void
    {
        BotJob::dispatchAfterResponse($bot);
    }

    public function deleting(Bot $bot): void
    {
        match ($bot->type) {
            BotType::News => $bot->articles?->each(fn (Article $article) => $article->delete()),
            BotType::Search => $bot->serp?->delete(),
            BotType::Sitemap => $bot->sitemap?->delete(),
        };
    }
}
