<?php

namespace Tests\Feature\Controller;

use App\Models\User;
use Laravel\Sanctum\Sanctum;
use Tests\TestCase;

class ApiAuthTest extends TestCase
{
    public function test_guest_is_unauthorized()
    {
        $this->getJson(route('api.bots.index'))->assertUnauthorized();
    }

    public function test_token_without_worker_ability_is_forbidden()
    {
        Sanctum::actingAs(User::factory()->create(), ['other']);

        $this->getJson(route('api.bots.index'))->assertForbidden();
    }

    public function test_token_with_worker_ability_is_allowed()
    {
        Sanctum::actingAs(User::factory()->create(), ['worker']);

        $this->getJson(route('api.bots.index'))->assertOk();
    }
}
