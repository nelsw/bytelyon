<?php

namespace Tests\Feature\Settings;

use App\Models\Proxy;
use App\Models\User;
use Tests\TestCase;

class ProxyTest extends TestCase
{
    public function test_proxies_page_is_displayed(): void
    {
        $user = User::factory()->create();

        $this->actingAs($user)
            ->get(route('proxies.edit'))
            ->assertOk();
    }

    public function test_proxies_page_lists_the_users_proxies(): void
    {
        $user = User::factory()->create();
        $user->proxies()->create(Proxy::factory()->makeOne(['name' => 'My proxy'])->toArray());

        $this->actingAs($user)
            ->get(route('proxies.edit'))
            ->assertOk()
            ->assertInertia(fn ($page) => $page->has('proxies', 1)
                ->where('proxies.0.name', 'My proxy')
            );
    }

    public function test_guests_cannot_view_proxies(): void
    {
        $this->get(route('proxies.edit'))->assertRedirect(route('login'));
    }

    public function test_guests_cannot_add_a_proxy(): void
    {
        $this->post(route('proxies.store'), [])->assertRedirect(route('login'));
    }

    public function test_a_proxy_can_be_added(): void
    {
        $user = User::factory()->create();

        $response = $this
            ->actingAs($user)
            ->post(route('proxies.store'), [
                'name' => 'My proxy',
                'scheme' => 'socks5',
                'host' => 'proxy.example.com',
                'port' => 8080,
                'user' => 'proxy-user',
                'pass' => 'secret',
                'bypass' => 'localhost',
            ]);

        $response
            ->assertSessionHasNoErrors()
            ->assertRedirect(route('proxies.edit'));

        $proxy = $user->proxies()->sole();

        $this->assertSame('My proxy', $proxy->name);
        $this->assertSame('socks5', $proxy->scheme);
        $this->assertSame('proxy.example.com', $proxy->host);
        $this->assertSame(8080, $proxy->port);
        $this->assertSame('proxy-user', $proxy->user);
        $this->assertSame('secret', $proxy->pass);
        $this->assertSame('localhost', $proxy->bypass);
    }

    public function test_a_proxy_requires_a_name_protocol_and_server(): void
    {
        $user = User::factory()->create();

        $this->actingAs($user)
            ->post(route('proxies.store'), [])
            ->assertSessionHasErrors(['name', 'scheme', 'host'])
            ->assertSessionDoesntHaveErrors(['port', 'user', 'pass', 'bypass']);
    }

    public function test_a_proxy_can_be_added_with_only_the_required_fields(): void
    {
        $user = User::factory()->create();

        $response = $this
            ->actingAs($user)
            ->post(route('proxies.store'), [
                'name' => 'My proxy',
                'scheme' => 'http',
                'host' => 'proxy.example.com',
            ]);

        $response
            ->assertSessionHasNoErrors()
            ->assertRedirect(route('proxies.edit'));

        $proxy = $user->proxies()->sole();

        $this->assertSame('My proxy', $proxy->name);
        $this->assertSame('http', $proxy->scheme);
        $this->assertSame('proxy.example.com', $proxy->host);
        $this->assertNull($proxy->port);
        $this->assertNull($proxy->user);
        $this->assertNull($proxy->pass);
        $this->assertNull($proxy->bypass);
    }

    public function test_a_proxy_requires_a_valid_protocol(): void
    {
        $user = User::factory()->create();

        $this->actingAs($user)
            ->post(route('proxies.store'), [
                'name' => 'My proxy',
                'scheme' => 'ftp',
                'host' => 'proxy.example.com',
                'port' => 8080,
                'user' => 'proxy-user',
                'pass' => 'secret',
            ])
            ->assertSessionHasErrors('scheme');
    }

    public function test_a_proxy_can_be_deleted(): void
    {
        $user = User::factory()->create();
        $proxy = $user->proxies()->create(Proxy::factory()->makeOne()->toArray());

        $this->actingAs($user)
            ->delete(route('proxies.destroy', $proxy->id))
            ->assertRedirect(route('proxies.edit'));

        $this->assertSame(0, $user->proxies()->count());
    }

    public function test_a_proxy_belonging_to_another_user_cannot_be_deleted(): void
    {
        $user = User::factory()->create();
        $other = User::factory()->create();
        $proxy = $other->proxies()->create(Proxy::factory()->makeOne()->toArray());

        $this->actingAs($user)
            ->delete(route('proxies.destroy', $proxy->id))
            ->assertForbidden();

        $this->assertSame(1, $other->proxies()->count());
    }
}
