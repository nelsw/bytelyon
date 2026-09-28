<?php

namespace Tests\Feature\Services;

use App\Exceptions\ShopifyException;
use App\Models\Shopify;
use App\Services\ShopifyService;
use Illuminate\Http\Client\ConnectionException;
use Illuminate\Http\Client\Request;
use Illuminate\Support\Facades\Cache;
use Illuminate\Support\Facades\Http;
use Tests\TestCase;

class ShopifyServiceTest extends TestCase
{
    private Shopify $shopify;

    protected function setUp(): void
    {
        parent::setUp();

        $this->shopify = new Shopify([
            'store' => 'store-'.fake()->uuid(),
            'client_id' => 'client-id',
            'client_secret' => 'client-secret',
            'default_author_name' => 'Jane',
            'default_blog_id' => '123',
        ]);
        Cache::forget("shopify:token:{$this->shopify->store}");
    }

    public function test_create_article(): void
    {
        Http::fake([
            '*/admin/oauth/access_token' => Http::response(['access_token' => 'tok']),
            '*/graphql.json' => Http::response(['data' => ['articleCreate' => ['article' => ['handle' => 'x']]]]),
        ]);

        new ShopifyService()->createArticle(
            shopify: $this->shopify,
            body: '<p>Body</p>',
            title: 'Title',
            tags: ['a'],
            handle: 'title',
            imageAlt: 'alt',
            imageUrl: 'https://example.com/i.png',
        );

        Http::assertSent(fn (Request $request) => str_ends_with($request->url(), '/graphql.json')
            && $request->hasHeader('X-Shopify-Access-Token', 'tok')
            && $request['variables']['article']['blogId'] === 'gid://shopify/Blog/123'
            && $request['variables']['article']['author'] === ['name' => 'Jane']
            && $request['variables']['article']['handle'] === 'title'
            && $request['variables']['article']['image'] === ['altText' => 'alt', 'url' => 'https://example.com/i.png']);
    }

    public function test_create_article_throws_on_graphql_errors(): void
    {
        Http::fake([
            '*/admin/oauth/access_token' => Http::response(['access_token' => 'tok']),
            '*/graphql.json' => Http::response(['errors' => [['message' => 'nope']]]),
        ]);

        $this->expectException(ShopifyException::class);

        new ShopifyService()->createArticle($this->shopify, 'body', 'title');
    }

    public function test_create_article_throws_on_connection_failure(): void
    {
        Http::fake([
            '*/admin/oauth/access_token' => Http::response(['access_token' => 'tok']),
            '*/graphql.json' => fn () => throw new ConnectionException('down'),
        ]);

        $this->expectException(ShopifyException::class);

        new ShopifyService()->createArticle($this->shopify, 'body', 'title');
    }

    public function test_create_access_token_throws_on_connection_failure(): void
    {
        Http::fake(fn () => throw new ConnectionException('down'));

        $this->expectException(ShopifyException::class);

        new ShopifyService()->createAccessToken($this->shopify);
    }
}
