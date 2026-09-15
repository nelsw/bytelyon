<?php

namespace Tests\Feature\Jobs;

use App\Jobs\NewsBotJob;
use App\Models\Bot;
use Tests\TestCase;

class NewsBotJobTest extends TestCase
{
    public function test_handle(): void
    {
        $bot = Bot::factory()->create();
        (new NewsBotJob($bot))->handle();
        $this->assertTrue(true);
    }
}
