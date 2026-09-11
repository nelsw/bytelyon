<?php

namespace Tests\Feature\Service;

use App\Models\Bot;
use App\Services\LambdaService;
use App\Services\NewsBotService;
use Tests\TestCase;

class NewsBotServiceTest extends TestCase
{
    private NewsBotService $service;

    protected function setUp(): void
    {
        parent::setUp();
        $this->service = resolve(NewsBotService::class, [
            'lambdaService' => resolve(LambdaService::class),
        ]);
    }

    public function test_run(): void
    {
        $bot = Bot::factory()
            ->news()
            ->enabled()
            ->query('btc forecast')
            ->lastRunAt(now()->subSecond())
            ->createOneQuietly();

        $this->assertDoesntThrow(fn () => $this->service->run($bot));
    }
}
