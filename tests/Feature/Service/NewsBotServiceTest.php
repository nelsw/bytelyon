<?php

namespace Tests\Feature\Service;

use App\Models\Bot;
use App\Services\LambdaService;
use App\Services\NewsBotService;
use Illuminate\Support\Facades\App;
use Tests\TestCase;

class NewsBotServiceTest extends TestCase
{
    private NewsBotService $service;

    protected function setUp(): void
    {
        parent::setUp();
        $this->service = resolve(NewsBotService::class, [
            'service' => resolve(LambdaService::class),
        ]);
    }

    public function test_run(): void
    {
        if (! App::hasDebugModeEnabled()) {
            return;
        }

        $bot = Bot::factory()
            ->news()
            ->enabled()
            ->query('btc forecast')
            ->lastRunAt(now()->subHours(6))
            ->createOneQuietly();

    }
}
