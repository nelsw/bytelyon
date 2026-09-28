<?php

namespace Tests\Unit\Console\Commands;

use App\Models\Bot;
use Illuminate\Bus\PendingBatch;
use Illuminate\Foundation\Testing\DatabaseTransactions;
use Illuminate\Support\Facades\Bus;
use Tests\TestCase;

class RunBotsTest extends TestCase
{
    use DatabaseTransactions;

    public function test_dispatches_a_batch_of_runnable_bots(): void
    {
        Bus::fake();
        Bot::factory()->news()->enabled()->neverRun()->createOneQuietly();

        $this->artisan('run:bots')->assertOk();

        Bus::assertBatched(fn (PendingBatch $batch) => $batch->name === 'run-bots' && $batch->jobs->isNotEmpty());
    }

    public function test_does_not_dispatch_when_no_bots_are_runnable(): void
    {
        Bus::fake();
        Bot::query()->update(['enabled' => false]);

        $this->artisan('run:bots')->expectsOutput('jobs to run [0]')->assertOk();

        Bus::assertBatchCount(0);
    }
}
