<?php

namespace Tests\Feature\Settings;

use App\Models\User;
use Tests\TestCase;

class IntegrationsTest extends TestCase
{
    public function test_integrations_page_is_displayed(): void
    {
        $user = User::factory()->create();

        $response = $this
            ->actingAs($user)
            ->get(route('integrations.edit'));

        $response->assertOk();
    }

    public function test_anthropic_settings_can_be_updated(): void
    {
        $user = User::factory()->create();

        $response = $this
            ->actingAs($user)
            ->put(route('integrations.anthropic.update'), [
                'api_key' => 'sk-ant-test-key',
                'default_model' => 'claude-sonnet-5',
            ]);

        $response
            ->assertSessionHasNoErrors()
            ->assertRedirect(route('integrations.edit'));

        $this->assertSame('sk-ant-test-key', $user->anthropic->api_key);
        $this->assertSame('claude-sonnet-5', $user->anthropic->default_model);
    }

    public function test_anthropic_settings_require_an_api_key(): void
    {
        $user = User::factory()->create();

        $response = $this
            ->actingAs($user)
            ->put(route('integrations.anthropic.update'), [
                'default_model' => 'claude-sonnet-5',
            ]);

        $response->assertSessionHasErrors('api_key');
    }

    public function test_shopify_settings_can_be_updated(): void
    {
        $user = User::factory()->create();

        $response = $this
            ->actingAs($user)
            ->put(route('integrations.shopify.update'), [
                'store' => 'my-store',
                'client_id' => 'client-id',
                'client_secret' => 'client-secret',
                'default_author_name' => 'Jane Doe',
                'default_blog_id' => '123',
            ]);

        $response
            ->assertSessionHasNoErrors()
            ->assertRedirect(route('integrations.edit'));

        $this->assertSame('my-store', $user->shopify->store);
        $this->assertSame('client-id', $user->shopify->client_id);
        $this->assertSame('client-secret', $user->shopify->client_secret);
        $this->assertSame('Jane Doe', $user->shopify->default_author_name);
        $this->assertSame('123', $user->shopify->default_blog_id);
    }

    public function test_shopify_settings_require_a_store_client_id_and_client_secret(): void
    {
        $user = User::factory()->create();

        $response = $this
            ->actingAs($user)
            ->put(route('integrations.shopify.update'), []);

        $response->assertSessionHasErrors(['store', 'client_id', 'client_secret']);
    }
}
