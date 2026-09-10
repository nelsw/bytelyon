<?php

namespace App\Services;

use App\Models\Bot;
use Illuminate\Container\Attributes\Singleton;
use Illuminate\Support\Facades\Log;
use Illuminate\Support\Facades\Storage;
use Illuminate\Support\Facades\URL;
use Illuminate\Support\Stringable;
use Uri\Rfc3986\Uri;

#[Singleton]
readonly class SitemapBotService
{
    public function __construct(private BotProcessService $botProcessService) {}

    public function run(Bot $bot): bool
    {
        Log::info("SitemapBotService - [$bot->query]");

        $url = "https://$bot->query";
        $map = [$url => false];

        $this->build($bot, 5, $map, [$url]);
        $bot->sitemap->update([
            'urls' => collect($map)->keys()->unique()->sort()->all(),
        ]);

        Log::info('SitemapBotService - completed', [
            'bot_id' => $bot->id,
            'domain' => $bot->query,
            'urls' => count($bot->sitemap->urls),
        ]);

        return true;
    }

    private function build(Bot $bot, int $depth, array &$done, array $next): void
    {
        $count = count($next);
        Log::debug("SitemapBotService - depth=[$depth] URLs=[$count]");
        if ($depth <= 0 || $count === 0) {
            return;
        }

        if (! $this->botProcessService->urls($bot, $next)) {
            Log::warning('SitemapBotService - processing error');
        }

        $todo = [];
        foreach ($next as $url) {

            $uuid = URL::toUuid5($url);
            $path = "sitemap/$bot->id/$uuid.json";

            $data = Storage::json($path);
            if ($data === null) {
                Log::warning("SitemapBotService - null file at [$path]");
                continue;
            }

            $page = $bot->sitemap->pages()->updateOrCreate(['url' => $url],
                [
                    'domain' => $bot->query,
                    'meta' => $data['meta'],
                    'screenshot_key' => "{$bot->type->value}/$bot->id/$uuid.png",
                    'title' => $data['title'],
                ]
            );
            Log::debug('SitemapBotService - saved', $page->toArray());

            $done[$url] = true;

            Log::debug("SitemapBotService - scraping [$url]");
            foreach ($data['links'] ?? [] as $link) {

                if (isset($done[$link]) ||
                    Uri::parse($link) === null ||
                    str_starts_with($link, 'http://') ||
                    URL::toDomain($link) !== $bot->query) {
                    continue;
                }

                $link = str($link)
                    ->trim()
                    ->before('#')
                    ->whenDoesntStartWith('https://', function (Stringable $str) {
                        return $str->prepend('https://');
                    })->toString();

                $done[$link] = false;
                $todo[] = $link;

                Log::debug("SitemapBotService - added [$link]");
            }
        }

        $this->build($bot, --$depth, $done, $todo);
    }
}
