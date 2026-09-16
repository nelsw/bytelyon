<?php

namespace Tests\Unit\Console\Commands;

use App\Models\User;
use Tests\TestCase;

class TestMintApiToken extends TestCase
{
    public function test_ok(): void
    {
        $this->artisan('api:token '.User::factory()->verified()->create()->email)->assertOk();
    }

    public function test_warn(): void
    {
        $this->assertThrows(fn () => $this->artisan('api:token'.fake()->email()));
    }
}
