<?php

namespace Tests\Feature\Controller;

use App\Models\Page;
use App\Models\Serp;
use App\Models\User;
use Inertia\Testing\AssertableInertia as Assert;
use Str;
use Tests\TestCase;

class SerpControllerTest extends TestCase
{
    public function test_authenticated_verified_users_can_view_available_serps(): void
    {
        $user = User::factory()->verified()->create();

        Serp::factory()->create();

        Serp::factory()->deleted()->create();

        $response = $this->actingAs($user)->get(route('serps.index'));

        $response->assertOk();

        $response->assertInertia(fn (Assert $page) => $page
            ->component('serps/Index')
            ->has('serps', 1)
        );
    }
}
