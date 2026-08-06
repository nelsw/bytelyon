<?php

namespace Tests\Feature\Controller;

use App\Models\Article;
use App\Models\Bot;
use App\Models\Page;
use App\Models\Serp;
use App\Models\Sitemap;
use Illuminate\Support\Facades\Redis;
use Tests\TestCase;

class ApiControllerTest extends TestCase
{
    private array $headers;

    protected function setUp(): void
    {
        parent::setUp();
        $this->headers = ['x-api-key' => config('app.whitelist.keys')];
    }

    public function test_bots_index()
    {
        $this->get(route('api.bots.index'), $this->headers)->assertOk();
    }

    public function test_bots_update()
    {
        $id = Bot::factory()->createQuietly()->id;
        $exp = fake()->sentence;

        $this->put(
            uri: route('api.bots.update', ['bot' => $id]),
            data: ['result' => $exp],
            headers: $this->headers,
        )->assertOk();

        $act = Redis::connection('broker')->getDel("bot:$id:done");

        $this->assertEquals($exp, $act);
    }

    public function test_article()
    {
        $model = Article::factory()->create();
        $this->put(
            route('api.articles.upsert', $model->bot),
            $model->toArray(),
            $this->headers,
        )
            ->assertOk();

    }

    public function test_serp()
    {
        $model = Serp::factory()->create();
        $this->put(
            route('api.searches.upsert', $model->bot),
            $model->toArray(),
            $this->headers,
        )->assertOk();
    }

    public function test_sitemap()
    {
        $this->put(
            route('api.sitemaps.upsert', Bot::factory()->createQuietly()),
            Sitemap::factory()->make(['urls' => null])->toArray(),
            $this->headers,
        )->assertOk();
    }

    public function test_sitemap_null_urls()
    {
        $model = Sitemap::factory()->create(['urls' => null]);
        $this->put(
            route('api.sitemaps.upsert', $model->bot),
            $model->toArray(),
            $this->headers,
        )->assertOk();
    }

    public function test_serp_page()
    {
        $model = Serp::factory()->create();
        $this->put(
            route('api.searches.pages.upsert', $model),
            Page::factory()->withRelation(Serp::class, $model->id)->create()->toArray(),
            $this->headers,
        )->assertOk();
    }

    public function test_sitemap_page()
    {
        $model = Sitemap::factory()->create();
        $this->put(
            route('api.sitemaps.pages.upsert', $model),
            Page::factory()->withRelation(Sitemap::class, $model->id)->make()->toArray(),
            $this->headers,
        )->assertOk();
    }
}
