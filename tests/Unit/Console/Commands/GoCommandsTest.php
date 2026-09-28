<?php

namespace Tests\Unit\Console\Commands;

use App\Facades\Go;
use Closure;
use Tests\TestCase;

class GoCommandsTest extends TestCase
{
    public function test_pub(): void
    {
        Go::shouldReceive('pub')->once()->with('42');

        $this->artisan('go:pub', ['id' => 42])
            ->expectsOutput('Sending bot (42) to the go engine.')
            ->expectsOutput('Sent.')
            ->assertOk();
    }

    public function test_sub(): void
    {
        Go::shouldReceive('sub')->once()->andReturnUsing(fn (Closure $callback) => $callback('hello'));

        $this->artisan('go:sub')
            ->expectsOutput('Listening to the go engine for events...')
            ->expectsOutput("Received:\nhello")
            ->expectsOutput('Done listening.')
            ->assertOk();
    }
}
