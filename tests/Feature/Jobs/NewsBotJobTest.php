<?php

namespace Tests\Feature\Jobs;

use App\Facades\Sqs;
use App\Jobs\NewsBotJob;
use App\Models\Bot;
use Tests\TestCase;

class NewsBotJobTest extends TestCase
{
    public function test_handle(): void
    {
        // Explicit ->news() state: Bot::factory()->create() alone picks a
        // random BotType, and only ->news() guarantees a News-shaped query
        // (RssService::news() expects a topic, not a domain or sentence
        // meant for another bot type).
        $bot = Bot::factory()->news()->create();

        // RssService hits a real RSS feed and gracefully degrades to []
        // on any network failure (see App\Services\RssService::items()),
        // so the exact number of enqueued jobs here is environment-
        // dependent — just make sure handle() never makes a real AWS call.
        Sqs::shouldReceive('enqueueScrape')->zeroOrMoreTimes();

        (new NewsBotJob($bot))->handle();

        $this->assertNotNull($bot->refresh()->last_run_at);
    }
}
