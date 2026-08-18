<?php

namespace App\Services;

use App\Exceptions\ShopifyException;
use App\Models\Shopify;
use DateTimeInterface;
use Illuminate\Container\Attributes\Singleton;
use Illuminate\Http\Client\ConnectionException;
use Illuminate\Support\Facades\Cache;
use Illuminate\Support\Facades\Http;
use Illuminate\Support\Facades\Log;

#[Singleton]
readonly class ShopifyService
{
    /** @throws ShopifyException */
    private function getAccessToken(Shopify $shopify): string
    {
        return Cache::remember(
            key: "shopify:token:$shopify->store",
            ttl: now()->addDay(),
            callback: fn () => $this->createAccessToken($shopify),
        );
    }

    /** @throws ShopifyException */
    public function createAccessToken(Shopify $shopify): string
    {
        try {
            return Http::asForm()->post("https://$shopify->store.myshopify.com/admin/oauth/access_token", [
                'grant_type' => 'client_credentials',
                'client_id' => $shopify->client_id,
                'client_secret' => $shopify->client_secret,
            ])->json('access_token');
        } catch (ConnectionException $e) {
            throw new ShopifyException(previous: $e);
        }
    }

    /** @throws ShopifyException */
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
        $query = 'mutation CreateArticle($article: ArticleCreateInput!)
{
	articleCreate(article: $article) {
	    article { handle }
    }
}';

        $variables = [
            'article' => [
                'isPublished' => true,
                'blogId' => 'gid://shopify/Blog/'.($blogId ?? $shopify->default_blog_id),
                'author' => ['name' => $author ?? $shopify->default_author_name],
                'title' => $title,
                'body' => $body,
                'summary' => $summary ?? '',
                'publishDate' => $publishedAt ?? now(),
                'tags' => $tags ?? [],
            ],
        ];

        if ($handle !== null) {
            $variables['article']['handle'] = $handle;
        }

        if ($imageUrl !== null && $imageAlt !== null) {
            $variables['article']['image'] = [
                'altText' => $imageAlt,
                'url' => $imageUrl,
            ];
        }

        try {
            $json = Http::withHeader('X-Shopify-Access-Token', $this->getAccessToken($shopify))->post(
                url: "https://$shopify->store.myshopify.com/admin/api/2026-07/graphql.json",
                data: compact('query', 'variables'),
            )->json();
        } catch (ConnectionException $e) {
            throw new ShopifyException(previous: $e);
        }

        if (isset($json['errors'])) {
            Log::warning('ShopifyService::createArticle', $json);
            throw new ShopifyException("failed to create article for store: $shopify->store");
        }
    }
}
