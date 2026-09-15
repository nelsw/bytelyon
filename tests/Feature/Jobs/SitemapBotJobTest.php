<?php

namespace Tests\Feature\Jobs;

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

        (new SitemapBotJob($bot))->handle();

        $this->assertNotEmpty($bot->refresh()->sitemap->pages);
    }
}
