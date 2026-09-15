<?php

namespace Tests\Feature\Model;

use App\Models\Page;
use App\Models\Sitemap;
use Tests\TestCase;

class SitemapModelTest extends TestCase
{
    public function test_observer(): void
    {
        /** @var Sitemap $sitemap */
        $sitemap = Sitemap::factory()->createQuietly();

        $this->assertDatabaseHas($sitemap);
        foreach ($sitemap->pages as $page) {
            $this->assertDatabaseHas(Page::class, [
                'pageable_id' => $sitemap->id,
                'pageable_type' => Sitemap::class,
                'id' => $page->id,
            ]);
        }

        $sitemap->bot->delete();

        $this->assertDatabaseMissing($sitemap);
        $this->assertSoftDeleted($sitemap->pages);
    }
}
