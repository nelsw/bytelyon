<?php

namespace Tests\Feature\Controller;

use App\Models\Article;
use App\Models\Bot;
use App\Models\Page;
use App\Models\Serp;
use App\Models\Sitemap;
use App\Models\User;
use Illuminate\Support\Facades\Redis;
use Laravel\Sanctum\Sanctum;
use Tests\TestCase;

class ApiControllerTest extends TestCase
{
    protected function setUp(): void
    {
        parent::setUp();
        Sanctum::actingAs(User::factory()->create(), ['worker']);
    }

    public function test_bots_index()
    {
        $this->get(route('api.bots.index'))->assertOk();
    }

    public function test_article()
    {
        $model = Article::factory()->create();
        $this->put(
            route('api.articles.upsert', $model->bot),
            $model->toArray(),
        )
            ->assertOk();

    }

    public function test_serp()
    {
        $model = Serp::factory()->create();
        $this->put(
            route('api.searches.upsert', $model->bot),
            $model->toArray(),
        )->assertOk();
    }

    public function test_sitemap()
    {
        $this->put(
            route('api.sitemaps.upsert', Bot::factory()->createQuietly()),
            Sitemap::factory()->make(['urls' => null])->toArray(),
        )->assertOk();
    }

    public function test_sitemap_null_urls()
    {
        $model = Sitemap::factory()->create(['urls' => null]);
        $this->put(
            route('api.sitemaps.upsert', $model->bot),
            $model->toArray(),
        )->assertOk();
    }

    public function test_serp_page()
    {
        $model = Serp::factory()->create();
        $this->put(
            route('api.searches.pages.upsert', $model),
            Page::factory()->withRelation(Serp::class, $model->id)->create()->toArray(),
        )->assertOk();
    }

    public function test_sitemap_page()
    {
        $model = Sitemap::factory()->create();
        $this->put(
            route('api.sitemaps.pages.upsert', $model),
            Page::factory()->withRelation(Sitemap::class, $model->id)->make()->toArray(),
        )->assertOk();
    }
}
