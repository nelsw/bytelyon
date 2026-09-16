<?php

namespace Tests\Feature\Controller;

use App\Models\Article;
use App\Models\Bot;
use App\Models\User;
use Illuminate\Foundation\Http\Middleware\PreventRequestForgery;
use Illuminate\Foundation\Testing\RefreshDatabase;
use Inertia\Testing\AssertableInertia as Assert;
use Tests\TestCase;

class ArticleControllerTest extends TestCase
{
    use RefreshDatabase;

    public function test_index()
    {
        $bot = Bot::factory()->enabled()->news()->createOneQuietly();
        $article = Article::factory()->for($bot)->count(3)->createQuietly();

        $response = $this->actingAs($bot->user)
            ->get(route('articles.index', $bot));

        $response->assertStatus(200)
            ->assertInertia(fn (Assert $page) => $page
                ->component('articles/Index')
                ->has('bot')
                ->has('articles', 3)
            );
    }

    public function test_show()
    {
        $article = Article::factory()->createOneQuietly();

        $response = $this->actingAs($article->bot->user)
            ->get(route('articles.show', ['bot' => $article->bot, 'article' => $article]));

        $response->assertStatus(200)
            ->assertInertia(fn (Assert $page) => $page
                ->component('articles/Show')
                ->has('bot')
                ->has('article')
            );
    }

    public function test_edit()
    {
        $article = Article::factory()->createOneQuietly();

        $response = $this->actingAs($article->bot->user)
            ->get(route('articles.edit', ['bot' => $article->bot, 'article' => $article]));

        $response->assertStatus(200)
            ->assertInertia(fn (Assert $page) => $page
                ->component('articles/Edit')
                ->has('bot')
                ->has('article')
                ->has('botOptions')
            );
    }

    public function test_update()
    {
        $article = Article::factory()->createOneQuietly();
        $newData = [
            'body' => 'Updated body',
            'description' => 'Updated description',
            'img_alt' => 'Updated alt',
            'img_url' => 'https://example.com/image.jpg',
            'keywords' => ['keyword1', 'keyword2'],
            'url' => 'https://example.com/article',
            'published_at' => now()->toDateTimeString(),
            'source' => 'Updated source',
            'title' => 'Updated Title',
        ];

        $response = $this->actingAs($article->bot->user)
            ->withoutMiddleware(PreventRequestForgery::class)
            ->put(route('articles.update', ['bot' => $article->bot, 'article' => $article]), $newData);

        $response->assertSessionHasNoErrors()
            ->assertRedirect(route('articles.edit', ['bot' => $article->bot, 'article' => $article]));
        $this->assertDatabaseHas('articles', [
            'id' => $article->id,
            'title' => 'Updated Title',
            'url' => 'https://example.com/article',
        ]);
    }

    public function test_update_without_source_or_bot_id_still_saves()
    {
        $article = Article::factory()->createQuietly([
            'source' => 'Original source',
        ]);

        $newData = [
            'body' => 'Updated body',
            'description' => 'Updated description',
            'img_alt' => 'Updated alt',
            'img_url' => 'https://example.com/image.jpg',
            'keywords' => ['keyword1', 'keyword2'],
            'url' => 'https://example.com/article',
            'published_at' => now()->toDateTimeString(),
            'title' => 'Updated Title',
        ];

        $response = $this->actingAs($article->bot->user)
            ->withoutMiddleware(PreventRequestForgery::class)
            ->put(route('articles.update', ['bot' => $article->bot, 'article' => $article]), $newData);

        $response->assertSessionHasNoErrors()
            ->assertRedirect(route('articles.edit', ['bot' => $article->bot, 'article' => $article]));
        $this->assertDatabaseHas('articles', [
            'id' => $article->id,
            'title' => 'Updated Title',
            'url' => 'https://example.com/article',
            'source' => 'Original source',
        ]);
    }

    public function test_cannot_access_other_users_bot_articles()
    {
        $tob = Bot::factory()->createQuietly();
        $bot = Bot::factory()->news()->createOneQuietly();
        $a = Article::factory()->for($bot)->createOneQuietly();

        $this->actingAs($tob->user)
            ->get(route('articles.index', $a))
            ->assertForbidden();

        $this->actingAs($tob->user)
            ->get(route('articles.show', ['bot' => $bot, 'article' => $a]))
            ->assertForbidden();
    }
}
