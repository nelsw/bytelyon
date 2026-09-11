<?php

namespace Tests\Feature\Settings;

use App\Models\User;
use Tests\TestCase;

class ApiTokenTest extends TestCase
{
    public function test_api_tokens_page_is_displayed(): void
    {
        $user = User::factory()->create();

        $this->actingAs($user)
            ->get(route('api-tokens.edit'))
            ->assertOk();
    }

    public function test_guests_cannot_view_api_tokens(): void
    {
        $this->get(route('api-tokens.edit'))->assertRedirect(route('login'));
    }

    public function test_a_token_can_be_issued(): void
    {
        $user = User::factory()->create();

        $response = $this
            ->actingAs($user)
            ->post(route('api-tokens.store'), ['name' => 'My scraper']);

        $response
            ->assertSessionHasNoErrors()
            ->assertRedirect(route('api-tokens.edit'))
            ->assertSessionHas('plainTextToken');

        $token = $user->tokens()->sole();

        $this->assertSame('My scraper', $token->name);
        $this->assertSame(['worker'], $token->abilities);
    }

    public function test_a_token_requires_a_name(): void
    {
        $user = User::factory()->create();

        $this->actingAs($user)
            ->post(route('api-tokens.store'), [])
            ->assertSessionHasErrors('name');
    }

    public function test_multiple_tokens_can_be_issued(): void
    {
        $user = User::factory()->create();

        $this->actingAs($user)->post(route('api-tokens.store'), ['name' => 'First']);
        $this->actingAs($user)->post(route('api-tokens.store'), ['name' => 'Second']);

        $this->assertSame(2, $user->tokens()->count());
    }

    public function test_a_token_can_be_revoked(): void
    {
        $user = User::factory()->create();
        $token = $user->createToken('My scraper', ['worker'])->accessToken;

        $this->actingAs($user)
            ->delete(route('api-tokens.destroy', $token->id))
            ->assertRedirect(route('api-tokens.edit'));

        $this->assertSame(0, $user->tokens()->count());
    }

    public function test_a_token_belonging_to_another_user_cannot_be_revoked(): void
    {
        $user = User::factory()->create();
        $other = User::factory()->create();
        $token = $other->createToken('Their scraper', ['worker'])->accessToken;

        $this->actingAs($user)
            ->delete(route('api-tokens.destroy', $token->id))
            ->assertNotFound();

        $this->assertSame(1, $other->tokens()->count());
    }
}
