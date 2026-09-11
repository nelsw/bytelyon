<?php

namespace Tests\Feature\Service;

use App\Models\Bot;
use App\Services\LambdaService;
use App\Services\SitemapBotService;
use Illuminate\Support\Facades\App;
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
        if (! App::hasDebugModeEnabled()) {
            return;
        }
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
