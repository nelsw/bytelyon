<?php

namespace Tests\Feature\Controller;

use App\Models\Bot;
use App\Models\Sitemap;
use App\Models\User;
use Inertia\Testing\AssertableInertia as Assert;
use Tests\TestCase;

class SitemapControllerTest extends TestCase
{
    public function test_authenticated_verified_users_can_view_their_sitemaps(): void
    {
        $user = User::factory()->verified()->create();

        $bot = Bot::factory()->sitemap()->for($user)->createQuietly();
        $sitemap = $bot->sitemap;

        // Belongs to a different user -- must not show up in the list.
        Sitemap::factory()->createQuietly();

        // Soft-deleted -- must not show up in the list either.
        Sitemap::factory()->deleted()->createQuietly();

        $response = $this->actingAs($user)->get(route('sitemaps.index'));

        $response->assertOk();

        $response->assertInertia(fn (Assert $page) => $page
            ->component('sitemaps/Index')
            ->has('sitemaps', 1)
            ->where('sitemaps.0.id', $sitemap->id)
            ->where('sitemaps.0.domain', $sitemap->domain)
            ->has('sitemaps.0.bot.frequency')
            ->has('sitemaps.0.bot.enabled')
        );
    }

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
