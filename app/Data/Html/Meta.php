<?php

namespace App\Data\Html;

use App\Data\Entry;
use Dom\HTMLDocument;
use Dom\HTMLElement;
use Illuminate\Contracts\Support\Arrayable;
use Illuminate\Support\Collection;

readonly class Meta implements Arrayable
{

    const array DESCRIPTION_KEYS = ["description", "og:description", "twitter:description", "abstract"];
    const array IMG_ALT_KEYS = ["og:image:alt", "twitter:image:alt"];
    const array IMG_SRC_KEYS = ["image", "og:image", "og:image:secure_url", "twitter:image", "twitter:image:src"];
    const array KEYWORD_KEYS = ["article:tag", "news_keywords", "keywords"];
    const array TITLE_KEYS = ["twitter:title", "og:title"];
    const array URL_KEYS = ["twitter:url", "og:url"];

    public function __construct(
        public array  $all,
        public string $description,
        public string $imgAlt,
        public string $imgSrc,
        public array  $keywords,
        public string $title,
        public string $url,
    )
    {
    }

    public static function fromArray(array $arr): self
    {
        return new self(
            all: $arr['all'] ?? [],
            description: $arr['description'] ?? '',
            imgAlt: $arr['img_alt'] ?? '',
            imgSrc: $arr['img_src'] ?? '',
            keywords: $arr['keywords'] ?? [],
            title: $arr['title'] ?? '',
            url: $arr['url'] ?? '',
        );
    }

    public static function empty(): self
    {
        return new self(
            all: [],
            description: '',
            imgAlt: '',
            imgSrc: '',
            keywords: [],
            title: '',
            url: '',
        );
    }

    public static function make(string $html): self
    {

        $all = collect(HTMLDocument::createFromString($html)->querySelectorAll('meta'))
            ->map(fn(HTMLElement $el) => new Entry(
                key: $el->getAttribute('name') ?: $el->getAttribute('property'),
                val: $el->getAttribute('content'),
            ))
            ->reject(fn(Entry $entry) => $entry->isKeyEmpty() || $entry->isValEmpty())
            ->groupBy('key')
            ->transform(function (Collection $entries) {
                if ($entries->count() === 1) {
                    return $entries[0]->toItem();
                }
                $vals = collect();
                foreach ($entries as $entry) {
                    $vals = $vals->mergeRecursive($entry->val);
                }
                return [$entries[0]->key => $vals->unique()->first()];
            })
            ->collapse()
            ->all();

        $arr = ['all' => $all];

        foreach (self::DESCRIPTION_KEYS as $key) {
            if (isset($all[$key])) {
                $arr['description'] = $all[$key];
                break;
            }
        }

        foreach (self::IMG_ALT_KEYS as $key) {
            if (isset($all[$key])) {
                $arr['imgAlt'] = $all[$key];
                break;
            }
        }

        foreach (self::IMG_SRC_KEYS as $key) {
            if (isset($all[$key])) {
                $arr['imgSrc'] = $all[$key];
                break;
            }
        }

        foreach (self::TITLE_KEYS as $key) {
            if (isset($all[$key])) {
                $arr['title'] = $all[$key];
                break;
            }
        }

        foreach (self::KEYWORD_KEYS as $key) {
            if (isset($all[$key])) {
                $arr['keywords'][] = $all[$key];
            }

        }

        return self::fromArray($arr);
    }

    public function isEmpty(): bool
    {
        return empty($this->all);
    }

    public function toArray(): array
    {
        return [
            'all' => $this->all,
            'description' => $this->description,
            'imgAlt' => $this->imgAlt,
            'imgSrc' => $this->imgSrc,
            'keywords' => $this->keywords,
            'title' => $this->title,
            'url' => $this->url,
        ];
    }
}
