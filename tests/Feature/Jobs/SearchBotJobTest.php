<?php

namespace Tests\Feature\Jobs;

use App\Facades\Sqs;
use App\Jobs\SearchBotJob;
use App\Models\Bot;
use Tests\TestCase;

class SearchBotJobTest extends TestCase
{
    public function test_handle(): void
    {
        // Explicit ->search() state: Bot::factory()->create() alone picks a
        // random BotType, and only ->search() guarantees the Serp relation
        // SearchBotJob::handle() needs (via BotObserver::created()).
        $bot = Bot::factory()->search()->create();

        Sqs::shouldReceive('enqueueScrape')
            ->once()
            ->with('serp', $bot->serp->id, ['query' => $bot->query]);

        (new SearchBotJob($bot))->handle();

        $this->assertNotNull($bot->refresh()->last_run_at);
    }
}
