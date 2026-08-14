<?php

namespace Tests\Feature\Controller;

use App\Models\Sitemap;
use App\Models\User;
use Inertia\Testing\AssertableInertia as Assert;
use Tests\TestCase;

class SitemapControllerTest extends TestCase
{
    public function test_authenticated_verified_users_can_view_available_sitemaps_without_urls(): void
    {
        Sitemap::query()->delete();

        $user = User::factory()->verified()->create();

        $sitemap = Sitemap::factory()->create();

        Sitemap::factory()->deleted()->create();

        $response = $this->actingAs($user)->get(route('sitemaps.index'));

        $response->assertOk();

        $response->assertInertia(fn (Assert $page) => $page
            ->component('sitemaps/Index')
            ->has('sitemaps', 1)
            ->where('sitemaps.0.domain', $sitemap->domain)
            ->where('sitemaps.0.deleted_at', null)
            ->has('sitemaps.0.urls', count($sitemap->urls))
        );
    }

    public function test_authenticated_verified_users_can_delete_a_sitemap(): void
    {
        $user = User::factory()->verified()->create();

        $sitemap = Sitemap::factory()->create();

        $response = $this->actingAs($user)
            ->delete(route('sitemaps.destroy', $sitemap));

        $response->assertRedirect(route('sitemaps.index'));
        $this->assertSoftDeleted('sitemaps', ['id' => $sitemap->id]);
    }

    public function test_authenticated_verified_users_can_view_sitemap_urls_as_tree_data(): void
    {
        $user = User::factory()->verified()->create();

        $sitemap = Sitemap::factory()->create();

        $response = $this->actingAs($user)
            ->get(route('sitemaps.show', $sitemap));

        $response->assertOk();

        $response->assertInertia(fn (Assert $page) => $page
            ->component('sitemaps/Show')
            ->where('sitemap.id', $sitemap->id)
            ->where('sitemap.domain', $sitemap->domain)
            ->where('sitemap.urls', $sitemap->urls)
        );
    }
}
