<?php

namespace Tests\Feature\Service;

use App\Models\Bot;
use App\Models\Sitemap;
use App\Services\LambdaService;
use App\Services\NewsBotService;
use App\Services\SitemapBotService;
use Tests\TestCase;

class SitemapBotServiceTest extends TestCase
{
    private SitemapBotService $service;

    protected function setUp(): void
    {
        parent::setUp();
        $this->service = resolve(SitemapBotService::class, [
            'lambdaService' => resolve(LambdaService::class),
        ]);
    }

    public function test_run(): void
    {
        $bot = Bot::factory()
            ->sitemap()
            ->enabled()
            ->query('bytelyon.com')
            ->lastRunAt(now()->subYear())
            ->createOneQuietly();

        $bot->sitemap()->create(['domain' => $bot->query]);

        $this->assertDoesntThrow(fn () => $this->service->run($bot));
    }
}
