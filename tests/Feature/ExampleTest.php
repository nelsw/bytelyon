<?php

namespace Tests\Feature;

use Tests\TestCase;

class ExampleTest extends TestCase
{
    public function test_returns_a_successful_response(): void
    {
        $this->get(route('home'))->assertOk();
    }

    public function test_privacy_policy_is_public(): void
    {
        $this->get(route('privacy'))
            ->assertOk()
            ->assertInertia(fn ($page) => $page->component('Privacy'));
    }
}
