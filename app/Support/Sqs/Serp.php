<?php

namespace App\Support\Sqs;

use Dom\Element;
use Illuminate\Support\Collection;
use Illuminate\Support\Uri;

final class Serp extends Page
{
    private const array BLOCK_PHRASES = [
        'unusual traffic',
        'our systems have detected',
    ];

    public function blocked(): bool
    {
        return str_contains($this->url, '/sorry/')
            || str($this->content)->contains(self::BLOCK_PHRASES, true);
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
        return collect($this->elements('[data-pcu]'))
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
                    'domain' => Uri::domain($url),
                ];
            })
            ->all();
    }

    /** @return array<int, array<string, mixed>> */
    private function organicProducts(): array
    {
        return collect($this->elements('product-viewer-entrypoint'))
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
        return collect($this->elements('[data-dtld]'))
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
        return collect($this->elements('h3[id]'))
            ->values()
            ->map(function (Element $e, int $i) {
                $parent = $e->parentElement;
                $url = $this->resolveUrl($parent?->getAttribute('href') ?? '');

                return [
                    'kind' => 'organic_result',
                    'index' => $i,
                    'title' => trim($e->textContent ?? ''),
                    'url' => $url,
                    'domain' => Uri::domain($url),
                ];
            })
            ->all();
    }

    /** @return array<int, array<string, mixed>> */
    private function similarQueries(): array
    {
        /** @var Collection<int, string> $values */
        $values = collect($this->elements('div[data-notify-expansion]'))
            ->map(fn (Element $e) => trim($e->getAttribute('data-q') ?? ''))
            ->filter(fn (string $q) => strlen($q) > 4);

        $botstuff = $this->element('div#botstuff');
        if ($botstuff !== null) {
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
}
