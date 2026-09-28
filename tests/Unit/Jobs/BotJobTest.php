<?php

namespace Tests\Unit\Jobs;

use App\Facades\Go;
use App\Jobs\BotJob;
use App\Models\Bot;
use Illuminate\Support\Facades\Log;
use RuntimeException;
use Tests\TestCase;

class BotJobTest extends TestCase
{
    public function test_handle_publishes_the_bot(): void
    {
        $bot = Bot::factory()->news()->createOneQuietly();
        Go::shouldReceive('pub')->once()->withArgs(fn (Bot $b) => $b->is($bot));

        $job = new BotJob($bot);
        $job->handle();

        $this->assertSame("bot:$bot->id", $job->uniqueId());
    }

    public function test_failed_logs_the_exception(): void
    {
        Log::spy();

        new BotJob(Bot::factory()->news()->makeOne())->failed(new RuntimeException('boom'));

        Log::shouldHaveReceived('error')->once()->withArgs(fn (string $message) => $message === 'BotJob');
    }
}
