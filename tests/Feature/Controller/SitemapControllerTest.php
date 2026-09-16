<?php

namespace Tests\Feature\Controller;

use App\Models\Sitemap;
use Inertia\Testing\AssertableInertia as Assert;
use Tests\TestCase;

class SitemapControllerTest extends TestCase
{

    public function test_authenticated_verified_users_can_delete_a_sitemap(): void
    {
        $sitemap = Sitemap::factory()->createQuietly();

        $response = $this->actingAs($sitemap->bot->user)
            ->delete(route('sitemaps.destroy', $sitemap));

        $response->assertRedirect(route('sitemaps.index'));
        $this->assertSoftDeleted('sitemaps', ['id' => $sitemap->id]);
    }

    public function test_authenticated_verified_users_can_view_sitemap_urls_as_tree_data(): void
    {
        $sitemap = Sitemap::factory()->createQuietly();

        $response = $this->actingAs($sitemap->bot->user)
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
