<?php

namespace Tests\Unit\Console\Commands;

use App\Models\User;
use Tests\TestCase;

class MintApiTokenTest extends TestCase
{
    public function test_ok(): void
    {
        $user = User::factory()->verified()->create();

        $this->artisan('api:token', ['email' => $user->email, '--name' => 'scraper'])->assertOk();

        $this->assertSame(['worker'], $user->tokens()->sole()->abilities);
    }

    public function test_fails_when_user_is_missing(): void
    {
        $this->artisan('api:token', ['email' => 'missing-'.fake()->uuid().'@example.com'])->assertFailed();
    }
}
