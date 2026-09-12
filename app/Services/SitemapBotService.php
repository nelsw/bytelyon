<?php

namespace App\Services;

use App\Events\BotResultsPersisted;
use App\Models\Bot;
use App\Models\Sitemap;
use Illuminate\Container\Attributes\Singleton;
use Illuminate\Support\Facades\Log;
use Throwable;

#[Singleton]
readonly class SitemapBotService
{
    public function __construct(private LambdaService $service) {}

    public function run(Bot $bot): void
    {
        Log::info("SitemapBotService::run - domain=[$bot->query]");

        $url = "https://$bot->query";
        $urls = [];

        $this->sync($bot, 5, $urls, $url);

        $bot->sitemap()->update(['urls' => $urls]);

        Log::info('SitemapBotService - completed', [
            'bot_id' => $bot->id,
            'domain' => $bot->query,
            'urls' => count($urls),
        ]);

        if (count($urls) > 0) {
            BotResultsPersisted::dispatch($bot, __(':count page(s) crawled for ":domain".', [
                'count' => count($urls),
                'domain' => $bot->query,
            ]));
        }
    }

    public function sync(Bot $bot, int $depth, array &$urls, string $url): void
    {
        $data = $this->service->page($url);

        try {
            $bot->sitemap->pages()->updateOrCreate(
                attributes: [
                    'url' => $url,
                    'pageable_type' => Sitemap::class,
                    'pageable_id' => $bot->sitemap->id,
                ],
                values: [
                    'domain' => $bot->query,
                    'title' => $data['title'],
                    'screenshot_key' => $data['screenshot_key'],
                    'meta' => $data['meta'],
                ],
            );
            $urls[$url] = true;
        } catch (Throwable $e) {
            Log::error('SitemapBotService - error scraping page', [
                'bot_id' => $bot->id,
                'domain' => $bot->query,
                'url' => $url,
                'error' => $e,
            ]);
        }

        foreach ($data['links'] as $link) {
            if ($urls[$link] ?? false) {
                continue;
            }
            $urls[$link] = false;
            $this->sync($bot, $depth - 1, $urls, $link);
        }
    }
}
