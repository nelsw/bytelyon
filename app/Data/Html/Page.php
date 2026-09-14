<?php

namespace App\Data\Html;

use App\Contracts\Pageable;
use Dom\Element;
use Dom\HTMLDocument;
use Illuminate\Support\Collection;
use Illuminate\Support\Facades\Log;
use Illuminate\Support\Facades\URL;
use Illuminate\Support\Uri;
use Throwable;

class Page implements Pageable
{
    private const array descriptionMetaKeys = [
        "name='description'",
        "property='og:description'",
        "name='twitter:description'",
        "name='abstract'",
    ];

    private const array keywordMetaKeys = ["property='article:tag'", "name='news_keywords'", "name='keywords'"];

    private const array imgAltMetaKeys = ["property='og:image:alt'", "name='twitter:image:alt'"];

    private const array imgSrcMetaKeys = [
        "name='image'",
        "property='og:image'",
        "property='og:image:secure_url'",
        "name='twitter:image'",
        "name='twitter:image:src'",
    ];

    private readonly HTMLDocument $doc;

    public function __construct(private readonly string $url, string $html)
    {
        try {
            $this->doc = HTMLDocument::createFromString($html);
        } catch (Throwable $e) {
            Log::error("Page - {$e->getMessage()}", [
                'url' => $this->url,
                'exception' => $e,
            ]);
            $this->doc = HTMLDocument::createEmpty();
        }
    }

    /**
     * @param  array<string>  $attrs
     */
    private function metaVal(array $attrs, bool $allVals = false): string
    {
        /** @var array<string> $arr */
        $arr = [];
        foreach ($attrs as $attr) {

            $val = trim($this->doc
                ->querySelector("meta[$attr]")
                ?->getAttribute('content') ?? '');

            if (empty($val)) {
                continue;
            }

            if (! $allVals) {
                return $val;
            }

            $arr = [
                ...$arr,
                ...explode(',', $val),
            ];
        }

        /** @var Collection<int, string> $c */
        $c = collect($arr);

        return $c
            ->transform(fn (string $val): string => trim($val))
            ->filter()
            ->join(',');
    }

    /** @return array<string, string> */
    public function meta(): array
    {
        $meta = [];

        $tags = $this->doc->querySelectorAll('meta');
        foreach ($tags as $tag) {
            $val = trim($tag->getAttribute('content') ?? '');
            if (empty($val)) {
                continue;
            }

            $key = trim($tag->getAttribute('name') ?? '');
            if (empty($key)) {
                $key = trim($tag->getAttribute('property') ?? '');
                if (empty($key)) {
                    continue;
                }
            }

            $meta[$key] = $val;
        }

        return $meta;
    }

    public function body(): string
    {
        foreach (['article', 'main', 'body', 'html'] as $tag) {
            if ($this->doc->querySelector($tag)) {
                return collect($this->doc->querySelector($tag)->querySelectorAll('p'))
                    ->transform(fn (Element $e): string => trim($e->textContent ?? ''))
                    ->reject(fn (string $href): bool => empty($href))
                    ->join(' ');
            }
        }

        return ''; // @codeCoverageIgnore
    }

    /**
     * @return array<string>
     */
    public function links(): array
    {
        return collect($this->doc->querySelectorAll('a'))
            ->transform(fn (Element $e): string => trim($e->getAttribute('href') ?? ''))
            ->reject(fn (string $href): bool => empty($href))
            ->map(fn (string $href): Uri => Uri::of($href))
            ->filter(fn (Uri $uri): bool => str($uri->host())->replace('www.', '')->is($this->domain()))
            ->transform(fn (Uri $uri): string => $uri->toString())
            ->toArray();
    }

    public function title(): string
    {
        return $this->doc->title;
    }

    public function url(): string
    {
        return $this->url;
    }

    public function domain(): string
    {
        return URL::toDomain($this->url);
    }

    public function description(): string
    {
        return $this->metaVal(self::descriptionMetaKeys);
    }

    public function imgAlt(): string
    {
        return $this->metaVal(self::imgAltMetaKeys);
    }

    public function imgSrc(): string
    {
        return $this->metaVal(self::imgSrcMetaKeys);
    }

    /**
     * @return array<string>
     */
    public function keywords(): array
    {
        return explode(',', $this->metaVal(self::keywordMetaKeys, true));
    }

    public function toArray(): array
    {
        return [
            'body' => $this->body(),
            'description' => $this->description(),
            'domain' => $this->domain(),
            'img_alt' => $this->imgAlt(),
            'img_src' => $this->imgSrc(),
            'keywords' => $this->keywords(),
            'title' => $this->title(),
            'url' => $this->url(),
        ];
    }
}
