<?php

namespace Tests\Feature\Service;

use App\Models\Bot;
use App\Services\LambdaService;
use App\Services\SearchBotService;
use Illuminate\Support\Facades\App;
use Tests\TestCase;

class SearchBotServiceTest extends TestCase
{
    private SearchBotService $service;

    protected function setUp(): void
    {
        parent::setUp();
        $this->service = resolve(SearchBotService::class, [
            'lambdaService' => resolve(LambdaService::class),
        ]);
    }

    public function test_run(): void
    {
        if (! App::hasDebugModeEnabled()) {
            return;
        }

        $bot = Bot::factory()
            ->sitemap()
            ->enabled()
            ->query('sailing blocks')
            ->lastRunAt(now()->subYear())
            ->createOneQuietly();

        $bot->serp()->create(['query' => $bot->query]);

        $bot->user->proxies()->create(config('services.proxy'));

    }
}
