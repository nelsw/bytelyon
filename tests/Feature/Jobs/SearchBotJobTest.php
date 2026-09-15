<?php

namespace Tests\Feature\Jobs;

use App\Jobs\SearchBotJob;
use App\Models\Bot;
use Tests\TestCase;

class SearchBotJobTest extends TestCase
{
    public function test_handle(): void
    {
        $bot = Bot::factory()->create();
        (new SearchBotJob($bot))->handle();
        $this->assertTrue(true);
    }
}
