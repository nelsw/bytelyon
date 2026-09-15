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
        $bot = Bot::factory()->news()->create();

        Sqs::shouldReceive('enqueueScrape')->zeroOrMoreTimes();

        (new NewsBotJob($bot))->handle();

        $this->assertNotNull($bot->refresh()->last_run_at);
    }
}
