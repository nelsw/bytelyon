<?php

namespace Tests\Feature\Model;

use App\Models\Sitemap;
use Tests\TestCase;

class SitemapModelTest extends TestCase
{
    public function test_observer(): void
    {
        /** @var Sitemap $sitemap */
        $sitemap = Sitemap::factory()->hasPages(3)->create();

        $this->assertDatabaseHas($sitemap);
        $this->assertDatabaseHas($sitemap->pages);

        $sitemap->bot()->delete();

        $this->assertSoftDeleted($sitemap->pages);
        $this->assertSoftDeleted($sitemap);
    }
}
