<?php

namespace App\Traits;

trait HasMeta
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

    /**
     * @param  array<string>  $metaKeys
     */
    private function metaVal(array $metaKeys, bool $allVals = false): string
    {
        /** @var array<string> $arr */
        $arr = [];
        foreach ($metaKeys as $key) {
            $attr = $this->attr("meta[$key]", 'content')->trim();
            if ($attr->isEmpty()) {
                continue;
            }
            if (! $allVals) {
                return $attr->value();
            }
            $arr = [...$arr, ...$attr->explode(',')];
        }
        return collect($arr)
            ->transform(fn (string $val): string => trim($val))
            ->filter()
            ->join(',');
    }

    /** @return array<string, string> */
    public function meta(): array
    {
        $meta = [];
        foreach ($this->elements('meta') as $e) {

            $val = str($e->getAttribute('content'))->trim();
            if ($val->isEmpty()) {
                continue;
            }

            $key = value(function (?string $name, ?string $property): string {
                return str($name)->trim()->whenEmpty(fn ($str) => $str->append($property))->trim();
            }, $e->getAttribute('name'), $e->getAttribute('property'));

            if (! empty($key) && ! isset($meta[$key])) {
                $meta[$key] = $val->value();
            }
        }
        return $meta;
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
}
