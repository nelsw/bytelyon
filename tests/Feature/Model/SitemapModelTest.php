<?php

namespace Tests\Feature\Model;

use App\Models\Bot;
use App\Models\Page;
use App\Models\Sitemap;
use Tests\TestCase;

class SitemapModelTest extends TestCase
{
    public function test_observer(): void
    {
        $bot = Bot::factory()
            ->sitemap('bytelyon.com')
            ->enabled()
            ->createQuietly();
        $sitemap = $bot->sitemap;

        $this->assertDatabaseHas($sitemap);
        foreach ($sitemap->pages as $page) {
            $this->assertDatabaseHas(Page::class, [
                'pageable_id' => $page->pageable_id,
                'pageable_type' => Sitemap::class,
                'id' => $page->id,
            ]);
        }

        $bot->delete();

        $this->assertDatabaseMissing($sitemap);
        $this->assertSoftDeleted($sitemap->pages);
    }
}
