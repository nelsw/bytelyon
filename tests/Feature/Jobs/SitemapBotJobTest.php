<?php

namespace Tests\Feature\Jobs;

use App\Facades\Sqs;
use App\Jobs\SitemapBotJob;
use App\Models\Bot;
use Tests\TestCase;

class SitemapBotJobTest extends TestCase
{
    public function test_handle(): void
    {
        $bot = Bot::factory()
            ->sitemap('bytelyon.com')
            ->enabled()
            ->headless()
            ->neverRun()
            ->createOneQuietly();

        Sqs::shouldReceive('enqueueScrape')
            ->once()
            ->with('sitemap', $bot->id, ['url' => 'https://bytelyon.com', 'depth' => 5]);

        (new SitemapBotJob($bot))->handle();

        $this->assertNotNull($bot->refresh()->last_run_at);
    }
}
