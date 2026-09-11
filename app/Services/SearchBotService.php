<?php

namespace App\Services;

use App\Models\Bot;
use Illuminate\Container\Attributes\Singleton;
use Illuminate\Support\Facades\Log;

#[Singleton]
readonly class SearchBotService
{
    public function __construct(private LambdaService $service) {}

    public function run(Bot $bot): void
    {
        $attributes = $this->service->serp(
            query: $bot->query,
            proxy: $bot->user->proxies->random(),
        );

        $bot->serp->update($attributes);

        Log::info('SearchBotService#run', [
            'query' => $bot->query,
            'attributes' => $attributes,
        ]);
    }
}
