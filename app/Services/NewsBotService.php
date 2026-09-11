<?php

namespace App\Services;

use App\Models\Bot;
use Illuminate\Container\Attributes\Singleton;
use Illuminate\Support\Facades\Log;

#[Singleton]
readonly class NewsBotService
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

        if ($items->isEmpty()) {
            Log::info('NewsBotService#run', [
                'query' => $bot->query,
                'items' => 0,
            ]);
            return;
        }

        $pages = $this->lambdaService->news($items->keys()->all());

        foreach ($pages as $page) {
            $bot->articles()->updateOrCreate(
                attributes: ['url' => $page['url']],
                values: [
                    ...$items->get($page['url']),
                    ...$page,
                ]
            );
        }

        Log::info('NewsBotService::run', [
            'query' => $bot->query,
            'items' => $items->count(),
            'pages' => count($pages),
        ]);
    }
}
