<?php

namespace Tests\Unit\Policies;

use App\Models\Bot;
use App\Models\Proxy;
use App\Models\Sitemap;
use App\Models\User;
use App\Policies\BotPolicy;
use App\Policies\ProxyPolicy;
use App\Policies\SitemapPolicy;
use Tests\TestCase;

class PolicyTest extends TestCase
{
    private User $admin;

    private User $stranger;

    protected function setUp(): void
    {
        parent::setUp();

        $this->admin = User::factory()->make(['email' => 'admin@example.com']);
        $this->stranger = User::factory()->create();
        config(['app.admin' => $this->admin->email]);
    }

    public function test_bot_policy(): void
    {
        $policy = new BotPolicy;
        $bot = Bot::factory()->news()->createOneQuietly();

        $this->assertTrue($policy->create($this->stranger));

        $this->assertTrue($policy->view($bot->user, $bot));
        $this->assertTrue($policy->view($this->admin, $bot));
        $this->assertFalse($policy->view($this->stranger, $bot));

        $this->assertTrue($policy->update($bot->user, $bot));
        $this->assertFalse($policy->update($this->admin, $bot));

        $this->assertTrue($policy->delete($bot->user, $bot));
        $this->assertTrue($policy->delete($this->admin, $bot));
        $this->assertFalse($policy->delete($this->stranger, $bot));
    }

    public function test_proxy_policy(): void
    {
        $policy = new ProxyPolicy;
        $proxy = Proxy::factory()->create();

        $this->assertTrue($policy->viewAny($this->stranger));
        $this->assertTrue($policy->create($this->stranger));

        foreach (['view', 'update', 'delete'] as $ability) {
            $this->assertTrue($policy->{$ability}($proxy->user, $proxy));
            $this->assertFalse($policy->{$ability}($this->stranger, $proxy));
        }
    }

    public function test_sitemap_policy(): void
    {
        $policy = new SitemapPolicy;
        $sitemap = Sitemap::factory()->createOneQuietly();

        $this->assertTrue($policy->viewAny($this->admin));
        $this->assertFalse($policy->viewAny($this->stranger));
        $this->assertTrue($policy->create($this->stranger));

        foreach (['view', 'update', 'delete'] as $ability) {
            $this->assertTrue($policy->{$ability}($sitemap->bot->user, $sitemap));
            $this->assertTrue($policy->{$ability}($this->admin, $sitemap));
            $this->assertFalse($policy->{$ability}($this->stranger, $sitemap));
        }

        $this->assertFalse($policy->restore($this->admin, $sitemap));
        $this->assertFalse($policy->forceDelete($this->admin, $sitemap));
    }
}
