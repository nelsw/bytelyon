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
        private BotProcessService $botProcessService,
    ) {}

    public function run(Bot $bot): bool
    {
        $arr = collect()
            ->merge($this->bingRssService->fetch($bot))
            ->merge($this->googleRssService->fetch($bot));

        Log::info('NewsBotService', [
            'bot' => $bot->toPrettyJson(),
            'articles' => $arr->count(),
        ]);

        if ($arr->isEmpty()) {
            return true;
        }

        $result = $this->botProcessService->news($arr->keys()->all());
        if (is_string($result)) {
            Log::error("NewsBotService - error: $result");
            return false;
        }

        foreach ($result as $res) {
            $bot->articles()->updateOrCreate(
                ['url' => $res['url']],
                [...$arr->get($res['url']), ...$res]
            );
        }

        return true;
    }
}
