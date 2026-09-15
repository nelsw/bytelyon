<?php

namespace App\Support;

use Dom\Element;
use Dom\HTMLDocument;
use Illuminate\Support\Collection;
use Illuminate\Support\Uri;

/**
 * PHP port of the structural, best-effort Google SERP parser that used to
 * live in docker/lambda/serp-scraper/lambda_handler.py's `_extract_data`
 * (and, before this port, was duplicated into docker/lambda/serp-di too).
 * Moved into the main app so every scrape worker (serp, news, sitemap, ...)
 * can share the exact same minimal response contract — {url, screenshot_key,
 * content_key} — with zero per-job-type parsing logic living in Python.
 *
 * Structural rather than class-name-based on purpose: Google's SERP markup
 * is unversioned and uses largely obfuscated/rotating class names, so this
 * looks for long-lived structural hooks (attributes/tag names like
 * `[data-pcu]` or `h3[id]`) instead. Still inherently best-effort — expect
 * to revisit as Google's markup shifts.
 *
 * One deliberate behavior difference from the original Python version:
 * result/product links are Google `/url?q=...`-style redirects. The Python
 * version resolved these to their final destination with a real HTTP
 * request through the same browser/proxy session. This PHP port has no live
 * browser to do that with (the HTML is static, already fetched by a
 * worker), so `resolveUrl()` only decodes the destination already embedded
 * in the redirect's own query string — correct for Google's own `/url?q=`
 * format (the common case), but won't follow arbitrary JS-based redirects
 * or multi-hop redirect chains the way a live navigation would.
 */
readonly class Serp
{
    private const array BLOCK_PHRASES = [
        'unusual traffic',
        'our systems have detected',
    ];

    final public function __construct(
        public string $url,
        private HTMLDocument $doc,
    ) {}

    public static function of(string $url, ?string $content = null): static
    {
        return new static($url, rescue(
            fn () => HTMLDocument::createFromString($content ?? '', LIBXML_NOERROR | LIBXML_HTML_NOIMPLIED),
            HTMLDocument::createEmpty(),
        ));
    }

    /**
     * A CAPTCHA/block page is a "successful" fetch (no exception, no
     * network-error interstitial) but useless for extraction.
     */
    public function blocked(): bool
    {
        if (str_contains($this->url, '/sorry/')) {
            return true;
        }

        $html = strtolower($this->rawHtml());

        return collect(self::BLOCK_PHRASES)->contains(fn (string $phrase) => str_contains($html, $phrase));
    }

    /** @return array<string, array<int, array<string, mixed>>> */
    public function data(): array
    {
        if ($this->blocked()) {
            return $this->emptyData();
        }

        return [
            'sponsored_results' => $this->sponsoredResults(),
            'sponsored_products' => $this->sponsoredProducts(),
            'organic_results' => $this->organicResults(),
            'organic_products' => $this->organicProducts(),
            'similar_queries' => $this->similarQueries(),
        ];
    }

    /** @return array<string, array<int, array<string, mixed>>> */
    public function emptyData(): array
    {
        return [
            'sponsored_results' => [],
            'sponsored_products' => [],
            'organic_results' => [],
            'organic_products' => [],
            'similar_queries' => [],
        ];
    }

    /** @return array<int, array<string, mixed>> */
    private function sponsoredResults(): array
    {
        return collect($this->doc->querySelectorAll('[data-pcu]'))
            ->values()
            ->map(function (Element $e, int $i) {
                $spans = $e->querySelectorAll('span');
                $title = $spans->item(0)?->textContent ?? '';
                $rawBrand = trim($spans->item(1)?->textContent ?? '');
                $brand = trim(explode('https://', $rawBrand)[0]);
                $url = $this->resolveUrl($e->getAttribute('href') ?? '');

                return [
                    'kind' => 'sponsored_result',
                    'index' => $i,
                    'title' => trim($title),
                    'brand' => $brand,
                    'url' => $url,
                    'domain' => $this->toDomain($url),
                ];
            })
            ->all();
    }

    /** @return array<int, array<string, mixed>> */
    private function organicProducts(): array
    {
        return collect($this->doc->querySelectorAll('product-viewer-entrypoint'))
            ->values()
            ->map(function (Element $e, int $i) {
                $div = $e->querySelector('div');
                $img = $e->querySelector('img');
                $title = $div?->getAttribute('aria-label') ?? '';
                if (blank($title) && $img) {
                    $title = $img->getAttribute('alt') ?? '';
                }

                return [
                    'kind' => 'organic_product',
                    'index' => $i,
                    'title' => trim($title),
                    'image' => $img?->getAttribute('src') ?? '',
                ];
            })
            ->all();
    }

    /** @return array<int, array<string, mixed>> */
    private function sponsoredProducts(): array
    {
        return collect($this->doc->querySelectorAll('[data-dtld]'))
            ->values()
            ->map(function (Element $e, int $i) {
                $div = $e->querySelector('div.pla-unit-container');
                $a = $div?->querySelector('a.pla-unit-img-container-link');
                $titleEl = $div?->querySelector("div[data-call_grow_wiz_event='true']");
                $priceEl = $div?->querySelector('div[aria-label]');
                $brandEl = $div?->querySelector("span[role='text']");
                $img = $a?->querySelector('img');
                $url = $this->resolveUrl($a?->getAttribute('href') ?? '');

                return [
                    'kind' => 'sponsored_product',
                    'index' => $i,
                    'domain' => $e->getAttribute('data-dtld') ?? '',
                    'url' => $url,
                    'image' => $img?->getAttribute('src') ?? '',
                    'title' => trim($titleEl?->textContent ?? ''),
                    'price' => $priceEl?->getAttribute('aria-label') ?? '',
                    'brand' => trim($brandEl?->textContent ?? ''),
                ];
            })
            ->all();
    }

    /** @return array<int, array<string, mixed>> */
    private function organicResults(): array
    {
        return collect($this->doc->querySelectorAll('h3[id]'))
            ->values()
            ->map(function (Element $e, int $i) {
                $parent = $e->parentElement;
                $url = $this->resolveUrl($parent?->getAttribute('href') ?? '');

                return [
                    'kind' => 'organic_result',
                    'index' => $i,
                    'title' => trim($e->textContent ?? ''),
                    'url' => $url,
                    'domain' => $this->toDomain($url),
                ];
            })
            ->all();
    }

    /** @return array<int, array<string, mixed>> */
    private function similarQueries(): array
    {
        /** @var Collection<int, string> $values */
        $values = collect($this->doc->querySelectorAll('div[data-notify-expansion]'))
            ->map(fn (Element $e) => trim($e->getAttribute('data-q') ?? ''))
            ->filter(fn (string $q) => strlen($q) > 4);

        $botstuff = $this->doc->querySelector('div#botstuff');
        if ($botstuff) {
            $values = $values->merge(
                collect($botstuff->querySelectorAll('a'))
                    ->map(fn (Element $e) => trim($e->textContent ?? ''))
                    ->filter(fn (string $t) => strlen($t) > 4)
            );
        }

        return $values->values()
            ->map(fn (string $value, int $i) => ['kind' => 'similar_query', 'index' => $i, 'value' => $value])
            ->all();
    }

    /**
     * Normalizes a (possibly relative) Google result link and, for the
     * common `/url?q=<destination>` redirect format, decodes the real
     * destination straight out of the query string — no live request
     * needed. See this class's own docblock for why that's not a full
     * equivalent of the original Python version's live-redirect-following.
     */
    private function resolveUrl(string $href): string
    {
        if ($href === '') {
            return $href;
        }

        if (! str_starts_with($href, 'https://www.google.com')) {
            if (! str_starts_with($href, '/')) {
                $href = "/$href";
            }
            $href = "https://www.google.com$href";
        }

        $query = [];
        parse_str((string) parse_url($href, PHP_URL_QUERY), $query);
        if (isset($query['q']) && str_starts_with($query['q'], 'http')) {
            return $query['q'];
        }
        if (isset($query['url']) && str_starts_with($query['url'], 'http')) {
            return $query['url'];
        }

        return $href;
    }

    private function toDomain(string $url): string
    {
        if ($url === '') {
            return '';
        }

        return str(Uri::of($url)->host() ?? '')->replace('www.', '')->toString();
    }

    private function rawHtml(): string
    {
        return rescue(fn () => $this->doc->saveHtml() ?? '', '');
    }
}
