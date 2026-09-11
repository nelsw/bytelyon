<?php

namespace App\Services;

use App\Models\Bot;
use Illuminate\Container\Attributes\Singleton;
use Illuminate\Support\Facades\Log;

#[Singleton]
readonly class SitemapBotService
{
    public function __construct(private LambdaService $service) {}

    public function run(Bot $bot, bool $sync = false): void
    {
        Log::info("SitemapBotService::run - domain=[$bot->query] sync=[$sync]");

        $url = "https://$bot->query";
        $urls = [];

        $this->sync($bot, 5, $urls, $url);

        $bot->sitemap->update(['urls' => $urls]);

        Log::info('SitemapBotService - completed', [
            'bot_id' => $bot->id,
            'domain' => $bot->query,
            'urls' => count($urls),
        ]);
    }

    public function sync(Bot $bot, int $depth, array &$urls, string $url): void
    {
        $data = $this->service->page($url);

        $links = [];
        foreach ($data['links'] as $link) {
            if (!isset($urls[$link])) {
                $urls[$link] = false;
                $links[] = $link;
            }
        }
        unset($data['links']);

        $bot->sitemap->pages()->updateOrCreate(['url' => $url], ...$data);
        $urls[$url] = true;

        foreach ($links as $link) {
            $this->sync($bot, $depth - 1, $urls, $link);
        }
    }
}
