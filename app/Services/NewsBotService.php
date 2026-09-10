<?php

namespace App\Services;

use App\Models\Bot;
use Illuminate\Container\Attributes\Singleton;
use Illuminate\Support\Facades\Log;

#[Singleton]
final readonly class NewsBotService
{
    public function __construct(
        private GoogleRssService $googleRssService,
        private BingRssService $bingRssService,
        private LambdaService $lambdaService,
    ) {}

    public function run(Bot $bot): void
    {
        $items = collect()
            ->merge($this->bingRssService->fetch($bot))
            ->merge($this->googleRssService->fetch($bot));

        Log::info('NewsBotService::run', [
            'bot' => $bot->toPrettyJson(),
            'items' => $items->count(),
        ]);

        if ($items->isEmpty()) {
            return;
        }

        $pages = $this->lambdaService->news($items->keys()->all());

        Log::info('NewsBotService::run', [
            'bot' => $bot->toPrettyJson(),
            'items' => $items->count(),
            'pages' => count($pages),
        ]);

        foreach ($pages as $res) {
            $bot->articles()->updateOrCreate(
                ['url' => $res['url']],
                [...$items->get($res['url']), ...$res]
            );
        }
    }
}
