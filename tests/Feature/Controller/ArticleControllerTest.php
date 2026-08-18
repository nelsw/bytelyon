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

    protected User $user;

    protected Bot $bot;

    protected function setUp(): void
    {
        parent::setUp();

        $this->user = User::factory()->verified()->create();
        $this->bot = Bot::factory()->create(['user_id' => $this->user->id]);
    }

    public function test_index()
    {
        Article::factory()->count(3)->create(['bot_id' => $this->bot->id]);

        $response = $this->actingAs($this->user)
            ->get(route('articles.index', $this->bot));

        $response->assertStatus(200)
            ->assertInertia(fn (Assert $page) => $page
                ->component('articles/Index')
                ->has('bot')
                ->has('articles', 3)
            );
    }

    public function test_show()
    {
        $article = Article::factory()->create(['bot_id' => $this->bot->id]);

        $response = $this->actingAs($this->user)
            ->get(route('articles.show', ['bot' => $this->bot, 'article' => $article]));

        $response->assertStatus(200)
            ->assertInertia(fn (Assert $page) => $page
                ->component('articles/Show')
                ->has('bot')
                ->has('article')
            );
    }

    public function test_edit()
    {
        $article = Article::factory()->create(['bot_id' => $this->bot->id]);

        $response = $this->actingAs($this->user)
            ->get(route('articles.edit', ['bot' => $this->bot, 'article' => $article]));

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
        $article = Article::factory()->create(['bot_id' => $this->bot->id]);
        $newData = [
            'body' => 'Updated body',
            'bot_id' => $this->bot->id,
            'description' => 'Updated description',
            'img_alt' => 'Updated alt',
            'img_url' => 'https://example.com/image.jpg',
            'keywords' => ['keyword1', 'keyword2'],
            'url' => 'https://example.com/article',
            'published_at' => now()->toDateTimeString(),
            'source' => 'Updated source',
            'title' => 'Updated Title',
        ];

        $response = $this->actingAs($this->user)
            ->withoutMiddleware(PreventRequestForgery::class)
            ->put(route('articles.update', ['bot' => $this->bot, 'article' => $article]), $newData);

        $response->assertSessionHasNoErrors()
            ->assertRedirect(route('articles.edit', ['bot' => $this->bot, 'article' => $article]));
        $this->assertDatabaseHas('articles', [
            'id' => $article->id,
            'title' => 'Updated Title',
            'url' => 'https://example.com/article',
        ]);
    }

    public function test_update_without_source_or_bot_id_still_saves()
    {
        $article = Article::factory()->create([
            'bot_id' => $this->bot->id,
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

        $response = $this->actingAs($this->user)
            ->withoutMiddleware(PreventRequestForgery::class)
            ->put(route('articles.update', ['bot' => $this->bot, 'article' => $article]), $newData);

        $response->assertSessionHasNoErrors()
            ->assertRedirect(route('articles.edit', ['bot' => $this->bot, 'article' => $article]));
        $this->assertDatabaseHas('articles', [
            'id' => $article->id,
            'title' => 'Updated Title',
            'url' => 'https://example.com/article',
            'source' => 'Original source',
        ]);
    }

    public function test_cannot_access_other_users_bot_articles()
    {
        $otherUser = User::factory()->create();
        $otherBot = Bot::factory()->create(['user_id' => $otherUser->id]);
        $article = Article::factory()->create(['bot_id' => $otherBot->id]);

        $this->actingAs($this->user)
            ->get(route('articles.index', $otherBot))
            ->assertForbidden();

        $this->actingAs($this->user)
            ->get(route('articles.show', ['bot' => $otherBot, 'article' => $article]))
            ->assertForbidden();
    }
}
