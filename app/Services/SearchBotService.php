<?php

namespace App\Services;

use App\Enums\SerpPart;
use App\Models\Bot;
use Dom\Element;
use Dom\HTMLDocument;
use Illuminate\Container\Attributes\Singleton;
use Illuminate\Http\Client\ConnectionException;
use Illuminate\Http\Client\RequestException;
use Illuminate\Support\Facades\Cache;
use Illuminate\Support\Facades\Http;
use Illuminate\Support\Facades\Storage;
use Illuminate\Support\Facades\URL;

#[Singleton]
readonly class SearchBotService
{
    public function __construct(private BotProcessService $botProcessService) {}

    public function run(Bot $bot): bool
    {
        if (! $this->botProcessService->query($bot, $bot->query)) {
            return false;
        }

        $uuid = URL::toUuid5('https://www.google.com/search?q='.urlencode($bot->query));

        $src = Storage::get("{$bot->type->value}/$bot->id/$uuid.html");
        $doc = HTMLDocument::createFromString($src);
        $data = [
            SerpPart::SpoPro->value => $this->sponsoredProducts($doc),
            SerpPart::SpoRes->value => $this->sponsoredResults($doc),
            SerpPart::OrgPro->value => $this->organicProducts($doc),
            SerpPart::OrgRes->value => $this->organicResults($doc),
            SerpPart::SimQry->value => $this->similarQueries($doc),
        ];

        $bot->serp->update([
            'screenshot_key' => "{$bot->type->value}/$bot->id/$uuid.png",
            'content_key' => "{$bot->type->value}/$bot->id/$uuid.html",
            'data' => $data,
        ]);

        return true;
    }

    private function finalURL(string $link): string
    {
        if (! str_starts_with($link, 'https://www.google.com')) {
            if (! str_starts_with($link, '/')) {
                $link = "/$link";
            }
            $link = "https://www.google.com$link";
        }

        $key = 'gsearch:decoded:'.sha1($link);
        if (is_string($val = Cache::get($key))) {
            return $val;
        }

        try {
            $url = Http::get($link)->throw()->effectiveUri();
        } catch (ConnectionException|RequestException) {
            return $link;
        }

        $key = 'gsearch:decoded:'.sha1((string) $url);
        Cache::put($key, $url, now()->addMinutes(30));
        return $url;
    }

    private function sponsoredResults(HTMLDocument $doc): array
    {
        return collect($doc->querySelectorAll('[data-pcu]'))->map(function (Element $e, int $i): array {
            $url = $this->finalURL($e->getAttribute('href'));
            return [
                'kind' => SerpPart::SpoPro->value,
                'index' => $i,
                'title' => $e->querySelector('span')[0]->text(),
                'brand' => str($e->querySelector('span')[1]->text())->before('https://'),
                'url' => $url,
                'domain' => URL::toDomain($url),
            ];
        })->all();
    }

    private function organicProducts(HTMLDocument $doc): array
    {
        return collect($doc->querySelectorAll('product-viewer-entrypoint'))->map(function (Element $e, int $i): array {
            $title = $e->querySelector('div')->getAttribute('aria-label');
            $img = $e->querySelector('img');
            if ($title == '') {
                $img->getAttribute('alt');
            }
            return [
                'kind' => SerpPart::OrgPro->value,
                'index' => $i,
                'title' => $title,
                'image' => $img->getAttribute('src'),
            ];
        })->all();
    }

    private function sponsoredProducts(HTMLDocument $doc): array
    {
        return collect($doc->querySelectorAll('[data-dtld]'))->map(function (Element $e, int $i): array {
            $div = $e->querySelector('div.pla-unit-container');
            $a = $div->querySelector('a.pla-unit-img-container-link');
            return [
                'kind' => SerpPart::SpoPro->value,
                'index' => $i,
                'domain' => $e->getAttribute('data-dtld'),
                'url' => $this->finalURL($a->getAttribute('href')),
                'image' => $a->querySelector('img')->getAttribute('src'),
                'title' => $div->querySelector("div[data-call_grow_wiz_event='true']"),
                'price' => $div->querySelector('div[aria-label]'),
                'brand' => $div->querySelector("span[role='text']"),
            ];
        })->all();
    }

    private function organicResults(HTMLDocument $doc): array
    {
        return collect($doc->querySelectorAll('h3[id]'))->map(function (Element $e, int $i): array {
            $url = $this->finalURL($e->parentElement->getAttribute('href'));
            return [
                'kind' => SerpPart::OrgRes->value,
                'index' => $i,
                'title' => $e->textContent,
                'url' => $url,
                'domain' => URL::toDomain($url),
            ];
        })->all();
    }

    private function similarQueries(HTMLDocument $doc): array
    {
        $all = collect($doc->querySelectorAll('div[data-notify-expansion]'))
            ->map(fn (Element $e): string => $e->getAttribute('data-q'))
            ->filter(fn (string $q): bool => strlen($q) > 4);

        $arr = collect($doc->querySelector('div#botstuff')->querySelectorAll('a'))
            ->map(fn (Element $e): string => $e->textContent)
            ->filter(fn (string $q): bool => strlen($q) > 4);

        return $all->merge($arr)->map(fn (string $s, int $i): array => [
            'kind' => SerpPart::SimQry->value,
            'index' => $i,
            'value' => $s,
        ])->all();
    }
}
