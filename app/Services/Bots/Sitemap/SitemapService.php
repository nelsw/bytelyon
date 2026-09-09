<?php

namespace App\Services\Bots\Sitemap;

use App\Models\Sitemap;
use Illuminate\Support\Arr;
use Illuminate\Support\Facades\Log;
use Illuminate\Support\Facades\Storage;
use Illuminate\Support\Facades\URL;
use Illuminate\Support\Stringable;
use Uri\Rfc3986\Uri;

class SitemapService
{
    public function run(Sitemap $sitemap): bool
    {
        Log::info("SitemapService - [$sitemap->domain]");

        $map = [$sitemap->URL() => false];

        $this->build($sitemap, 5, $map, [$sitemap->URL()]);
        $sitemap->update([
            'urls' => collect($map)->keys()->unique()->sort()->all(),
        ]);

        Log::info("SitemapService - completed", [
            'bot_id' => $sitemap->id,
            'domain' => $sitemap->domain,
            'urls' => count($sitemap->urls),
        ]);

        return true;
    }

    private function build(Sitemap $sitemap, int $depth, array &$done, array $next): void
    {
        $count = count($next);
        Log::debug("SitemapService - depth=[$depth] URLs=[$count]");
        if ($depth <= 0 || $count === 0) {
            return;
        }

        if (!$sitemap->process(Arr::prepend($next, '-u'))) {
            Log::warning("SitemapService - processing error");
        }

        $todo = [];
        foreach ($next as $url) {

            $uuid = URL::toUuid5($url);
            $path = "sitemap/$sitemap->id/$uuid.json";

            $data = Storage::json($path);
            if ($data === null) {
                Log::warning("SitemapService - null file at [$path]");
                continue;
            }

            $page = $sitemap->pages()->updateOrCreate(['url' => $url],
                [
                    'domain' => $sitemap->domain,
                    'meta' => $data['meta'],
                    'screenshot_key' => "sitemap/$sitemap->id/$uuid.png",
                    'title' => $data['title'],
                ]
            );
            Log::debug("SitemapService - saved", $page->toArray());

            $done[$url] = true;

            Log::debug("SitemapService - scraping [$url]");
            foreach ($data['links'] ?? [] as $link) {

                if (isset($done[$link]) ||
                    Uri::parse($link) === null ||
                    str_starts_with($link, "http://") ||
                    URL::toDomain($link) !== $sitemap->domain) {
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

                Log::debug("SitemapService - added [$link]");
            }
        }

        $this->build($sitemap, --$depth, $done, $todo);
    }
}
