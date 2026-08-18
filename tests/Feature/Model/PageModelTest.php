<?php

namespace Tests\Feature\Model;

use App\Models\Page;
use App\Models\Serp;
use App\Models\Sitemap;
use Tests\TestCase;

class PageModelTest extends TestCase
{
    public function test_sitemap_pages(): void
    {
        $sitemap = Sitemap::factory()->create();

        $this->assertEmpty($sitemap->pages()->get());

        $count = 3;

        /** @var Page[] $pages */
        $pages = Page::factory()
            ->withRelation(Sitemap::class, $sitemap->getKey())
            ->count($count)
            ->create();

        $this->assertCount($count, $pages);

        foreach ($pages as $page) {
            $this->assertDatabaseHas('pages', [
                'id' => $page->getKey(),
                'pageable_id' => $sitemap->getKey(),
                'pageable_type' => $sitemap->getMorphClass(),
            ]);
        }

        $this->assertCount($count, $sitemap->pages()->get());
    }

    public function test_serp_pages(): void
    {
        $serp = Serp::factory()->create();

        $this->assertEmpty($serp->pages()->get());

        $count = 3;

        /** @var Page[] $pages */
        $pages = Page::factory()
            ->withRelation(Serp::class, $serp->getKey())
            ->count($count)
            ->create();

        $this->assertCount($count, $pages);

        foreach ($pages as $page) {
            $this->assertDatabaseHas('pages', [
                'id' => $page->getKey(),
                'pageable_id' => $serp->getKey(),
                'pageable_type' => $serp->getMorphClass(),
            ]);
        }

        $this->assertCount($count, $serp->pages()->get());
    }
}
