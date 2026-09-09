<?php

namespace App\Dto;

readonly class Meta
{
    private function __construct(
        public array $items,
        public string $description,
        public string $imageSrc,
        public string $imageAlt,
        public array $keywords,
    ) {}

    public static function of(array $items): self
    {
        return new self(
            items: collect($items)
                ->reject(fn ($item) => sizeof($item) < 2)
                ->transform(fn ($item) => [$item[0] => $item[1]])
                ->collapse()
                ->all(),
            description: self::find($items, fn ($item) => str_contains($item[0], 'description')),
            imageSrc: self::find($items, fn ($item) => $item[0] === 'image' || str_contains($item[0], ':image')),
            imageAlt: self::find($items, fn ($item) => str_contains($item[0], ':alt')),
            keywords: collect($items)
                ->filter(fn ($item) => str_contains($item[0], 'keyword'))
                ->transform(fn ($item) => explode(',', $item[1]))
                ->collapse()
                ->values()
                ->unique()
                ->sort()
                ->all(),
        );
    }
    private static function find($items, $callback): string
    {
        return collect($items)
            ->filter($callback)
            ->sortBy(fn ($item) => strlen($item[1]))
            ->values()
            ->first()[1] ?? '';
    }
}
