<?php

namespace App\Services;

use App\Data\Html\Page;
use App\Events\BotResultsPersisted;
use App\Facades\Lambda;
use App\Models\Bot;
use App\Models\Sitemap;
use Illuminate\Container\Attributes\Singleton;
use Illuminate\Support\Facades\Log;
use Illuminate\Support\Facades\Storage;
use Throwable;

#[Singleton]
readonly class SitemapBotService
{
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
        $payload = Lambda::scrape($url);

        $page = new Page($url, Storage::disk('s3')->get($payload->contentKey));

        try {
            $bot->sitemap->pages()->updateOrCreate(
                attributes: [
                    'url' => $url,
                    'pageable_type' => Sitemap::class,
                    'pageable_id' => $bot->sitemap->id,
                ],
                values: [
                    'domain' => $bot->query,
                    'title' => $page->title(),
                    'screenshot_key' => $payload->screenshotKey,
                    'meta' => $page->meta(),
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

        foreach ($page->links() as $link) {
            if ($urls[$link] ?? false) {
                continue;
            }
            $urls[$link] = false;
            $this->sync($bot, $depth - 1, $urls, $link);
        }
    }
}
