<?php

namespace Tests\Feature\Controller;

use Anthropic\Core\Exceptions\APIException;
use Anthropic\Messages\Model as AnthropicModel;
use App\Exceptions\ShopifyException;
use App\Models\Article;
use App\Models\Bot;
use App\Models\Shopify;
use App\Services\AnthropicService;
use App\Services\ShopifyService;
use Closure;
use DateTimeInterface;
use GuzzleHttp\Psr7\Request as Psr7Request;
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
            ->get(route('articles.index', $bot))
            ->assertForbidden();

        $this->actingAs($tob->user)
            ->get(route('articles.show', ['bot' => $bot, 'article' => $a]))
            ->assertForbidden();
    }

    public function test_update_is_forbidden_for_other_users()
    {
        $article = Article::factory()->createOneQuietly();
        $stranger = Bot::factory()->createOneQuietly()->user;

        $this->actingAs($stranger)
            ->put(route('articles.update', ['bot' => $article->bot, 'article' => $article]), ['title' => 'x'])
            ->assertForbidden();
    }

    public function test_destroy()
    {
        $article = Article::factory()->createOneQuietly();

        $this->actingAs($article->bot->user)
            ->delete(route('articles.destroy', ['bot' => $article->bot, 'article' => $article]))
            ->assertRedirect(route('articles.index', $article->bot));

        $this->assertSoftDeleted($article);
    }

    public function test_assist_requires_an_anthropic_key()
    {
        $article = Article::factory()->createOneQuietly();

        $this->actingAs($article->bot->user)
            ->postJson(route('articles.assist', ['bot' => $article->bot, 'article' => $article]), ['prompt' => 'Shorten it'])
            ->assertStatus(422)
            ->assertJsonStructure(['message']);
    }

    public function test_assist_returns_html()
    {
        $article = Article::factory()->createOneQuietly();
        $user = $article->bot->user;
        $user->anthropic()->create(['api_key' => 'sk-test', 'default_model' => 'claude-sonnet-5']);

        $calls = new \ArrayObject;
        $this->fakeAnthropic(function (array $args) use ($calls) {
            $calls[] = $args;

            return '**Shorter**';
        });

        $this->actingAs($user)
            ->postJson(route('articles.assist', ['bot' => $article->bot, 'article' => $article]), [
                'prompt' => 'Shorten it',
                'system' => '  Be brief.  ',
                'body' => '<p>Long</p>',
            ])
            ->assertOk()
            ->assertJson(['html' => "<p><strong>Shorter</strong></p>\n"]);

        $this->assertCount(1, $calls);
        $this->assertSame('sk-test', $calls[0]['apiKey']);
        $this->assertSame('claude-sonnet-5', $calls[0]['model']);
        $this->assertSame('Be brief.', $calls[0]['system']);
    }

    public function test_assist_reports_api_failures()
    {
        $article = Article::factory()->createOneQuietly();
        $user = $article->bot->user;
        $user->anthropic()->create(['api_key' => 'sk-test']);

        $this->fakeAnthropic(fn () => throw new APIException(new Psr7Request('POST', 'https://api.anthropic.com')));

        $this->actingAs($user)
            ->postJson(route('articles.assist', ['bot' => $article->bot, 'article' => $article]), ['prompt' => 'Shorten it'])
            ->assertStatus(502);
    }

    public function test_publish_requires_shopify_settings()
    {
        $article = Article::factory()->createOneQuietly();

        $calls = new \ArrayObject;
        $this->fakeShopify(fn (array $args) => $calls[] = $args);

        $this->actingAs($article->bot->user)
            ->post(route('articles.publish', ['bot' => $article->bot, 'article' => $article]))
            ->assertRedirect(route('articles.edit', ['bot' => $article->bot, 'article' => $article]));

        $this->assertCount(0, $calls);
    }

    public function test_publish()
    {
        $article = Article::factory()->createOneQuietly();
        $user = $article->bot->user;
        $user->shopify()->create(['store' => 'shop', 'client_id' => 'id', 'client_secret' => 'secret']);

        $calls = new \ArrayObject;
        $this->fakeShopify(fn (array $args) => $calls[] = $args);

        $this->actingAs($user)
            ->post(route('articles.publish', ['bot' => $article->bot, 'article' => $article]))
            ->assertRedirect(route('articles.edit', ['bot' => $article->bot, 'article' => $article]));

        $this->assertCount(1, $calls);
        $this->assertSame($article->title, $calls[0]['title']);
    }

    public function test_publish_handles_shopify_failures()
    {
        $article = Article::factory()->createOneQuietly();
        $user = $article->bot->user;
        $user->shopify()->create(['store' => 'shop', 'client_id' => 'id', 'client_secret' => 'secret']);

        $this->fakeShopify(fn () => throw new ShopifyException);

        $this->actingAs($user)
            ->post(route('articles.publish', ['bot' => $article->bot, 'article' => $article]))
            ->assertRedirect(route('articles.edit', ['bot' => $article->bot, 'article' => $article]));
    }

    private function fakeAnthropic(Closure $handler): void
    {
        $this->app->instance(AnthropicService::class, new readonly class($handler) extends AnthropicService
        {
            public function __construct(private Closure $handler) {}

            public function prompt(
                string $apiKey,
                array $messages,
                int $maxTokens = 1024,
                AnthropicModel|string $model = AnthropicModel::CLAUDE_FABLE_5,
                array|string|null $system = null,
                bool $html = false,
            ): string {
                return ($this->handler)(get_defined_vars());
            }
        });
    }

    private function fakeShopify(Closure $handler): void
    {
        $this->app->instance(ShopifyService::class, new readonly class($handler) extends ShopifyService
        {
            public function __construct(private Closure $handler) {}

            public function createArticle(
                Shopify $shopify,
                string $body,
                string $title,
                array $tags = [],
                ?DateTimeInterface $publishedAt = null,
                ?string $author = null,
                ?string $blogId = null,
                ?string $handle = null,
                ?string $imageAlt = null,
                ?string $imageUrl = null,
                ?string $summary = null,
            ): void {
                ($this->handler)(get_defined_vars());
            }
        });
    }
}
