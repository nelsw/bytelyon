<?php

namespace App\Services\Bots\Search;

use App\Enums\BotType;
use App\Enums\SerpPart;
use App\Models\Bot;
use Dom\Element;
use Dom\HTMLDocument;
use Illuminate\Container\Attributes\Singleton;
use Illuminate\Http\Client\ConnectionException;
use Illuminate\Http\Client\RequestException;
use Illuminate\Process\Exceptions\ProcessTimedOutException;
use Illuminate\Support\Facades\Cache;
use Illuminate\Support\Facades\Http;
use Illuminate\Support\Facades\Log;
use Illuminate\Support\Facades\Process;
use Illuminate\Support\Facades\Storage;
use Illuminate\Support\Facades\URL;
use RuntimeException;
use Throwable;

#[Singleton]
readonly class SearchBotService
{
    private const string HOST = "https://www.google.com";
    public function run(Bot $bot): bool
    {
        $args = [
            'uv', 'run', base_path('scripts/search.py'),
            '-p', implode('/', [storage_path('app/private'), $bot->type->value, $bot->id]),
            '-q', $bot->query
        ];
        if ($bot->headless) {
            $args[] = '--headless';
        }

        $cmd = implode(' ', $args);
        try {
            $result = Process::run($args, function (string $type, string $output) use ($args, $cmd) {
                Log::debug('SearchBotService::run', [
                    'cmd' => $cmd,
                    $type => $output,
                ]);
            });
        } catch (ProcessTimedOutException | RuntimeException $e) {
            Log::error('SearchBotService::run', [
                'cmd' => $cmd,
                'error' => $e->getMessage(),
                'ok' => false,
            ]);
            return false;
        }

        Log::info('SearchBotService::run', [
            'cmd' => $cmd,
            'ok' => $result->successful(),
        ]);

        if ($result->failed()) {
            return false;
        }

        // url is not actual page URL, but rather a simple
        // concatenation of parts to prevent uuid shifts
        $url = self::HOST . '/search?q=' . urlencode($bot->query);
        $uuid = URL::toUuid5($url);

        $path = "{$bot->type->value}/$bot->id/$uuid.html";
        $doc = HTMLDocument::createFromString(Storage::get($path));

        $data = [
            SerpPart::SpoPro->value => $this->sponsoredProducts($doc),
            SerpPart::SpoRes->value => $this->sponsoredResults($doc),
            SerpPart::OrgPro->value => $this->organicProducts($doc),
            SerpPart::OrgRes->value => $this->organicResults($doc),
            SerpPart::SimQry->value => $this->similarQueries($doc),
        ];

        return true;
    }

    private function finalURL(string $link): string
    {
        if (!str_starts_with($link, self::HOST)) {
            $link = self::HOST . $link;
        }

        $key = 'gsearch:decoded:' . sha1($link);
        if (is_string($val = Cache::get($key))) {
            return $val;
        }

        try {
            $url = Http::get($link)->throw()->effectiveUri();
        } catch (ConnectionException|RequestException) {
            return $link;
        }

        $key = 'gsearch:decoded:' . sha1((string)$url);
        Cache::put($key, $url, now()->addMinutes(30));
        return $url;
    }

    private function sponsoredResults(HTMLDocument $doc): array
    {
        return collect($doc->querySelectorAll("[data-pcu]"))->map(function (Element $e, int $i): array {
            $url = $this->finalURL($e->getAttribute('href'));
            return [
                "kind" => SerpPart::SpoPro->value,
                'index' => $i,
                'title' => $e->querySelector("span")[0]->text(),
                'brand' => str($e->querySelector("span")[1]->text())->before('https://'),
                'url' => $url,
                'domain' => URL::toDomain($url),
            ];
        })->all();
    }

    private function organicProducts(HTMLDocument $doc): array
    {
        return collect($doc->querySelectorAll("product-viewer-entrypoint"))->map(function (Element $e, int $i): array {
            $title = $e->querySelector("div")->getAttribute('aria-label');
            $img = $e->querySelector("img");
            if ($title == '') {
                $img->getAttribute("alt");
            }
            return [
                "kind" => SerpPart::OrgPro->value,
                'index' => $i,
                'title' => $title,
                'image' => $img->getAttribute('src'),
            ];
        })->all();
    }

    private function sponsoredProducts(HTMLDocument $doc): array
    {
        return collect($doc->querySelectorAll("[data-dtld]"))->map(function (Element $e, int $i): array {
            $div = $e->querySelector("div.pla-unit-container");
            $a = $div->querySelector("a.pla-unit-img-container-link");
            return [
                "kind" => SerpPart::SpoPro->value,
                "index" => $i,
                "domain" => $e->getAttribute("data-dtld"),
                "link" => $this->finalURL($a->getAttribute("href")),
                "image" => $a->querySelector("img")->getAttribute("src"),
                "title" => $div->querySelector("div[data-call_grow_wiz_event='true']"),
                "price" => $div->querySelector("div[aria-label]"),
                "brand" => $div->querySelector("span[role='text']"),
            ];
        })->all();
    }

    private function organicResults(HTMLDocument $doc): array
    {
        return collect($doc->querySelectorAll("h3[id]"))->map(function (Element $e, int $i): array {
            return [
                "kind" => SerpPart::OrgRes->value,
                "index" => $i,
                "title" => $e->textContent,
                "url" => $this->finalURL($e->parentElement->getAttribute("href")),
            ];
        })->all();
    }

    private function similarQueries(HTMLDocument $doc): array
    {
        $all = collect($doc->querySelectorAll("div[data-notify-expansion]"))
            ->map(fn(Element $e): string => $e->getAttribute("data-q"))
            ->filter(fn(string $q): bool => strlen($q) > 4);

        $arr = collect($doc->querySelector("div#botstuff")->querySelectorAll("a"))
            ->map(fn(Element $e): string => $e->textContent)
            ->filter(fn(string $q): bool => strlen($q) > 4);

        // Note: The concat method numerically re-indexes keys for items concatenated onto the original collection.
        // While merge returns a new collection, it also preserves the order of keys in associative collections.
        return $all->merge($arr)->map(fn(string $s, int $i): array => [
            "kind" => SerpPart::SimQry->value,
            "index" => $i,
            "value" => $s,
        ])->all();
    }
}
