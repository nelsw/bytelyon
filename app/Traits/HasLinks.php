<?php

namespace App\Traits;

use Dom\Element;
use Illuminate\Support\Collection;
use Illuminate\Support\Stringable;
use Illuminate\Support\Uri;

trait HasLinks
{
    /**
     * @return Collection<int, string>
     */
    public function links(): Collection
    {
        return $this
            ->anchorElements()
            ->transform(self::toHref())
            ->reject(self::deadEnds())
            ->transform(self::toUri($this->url))
            ->filter(self::sameSite($this->domain))
            ->transform(self::toString())
            ->unique()
            ->values();
    }

    private function anchorElements(): Collection
    {
        return collect($this->elements('a'));
    }

    private static function toHref(): callable
    {
        return function (Element $element): Stringable {
            return str($element->getAttribute('href'))->trim();
        };
    }

    private static function deadEnds(): callable
    {
        return function (Stringable $href): bool {
            return $href->isEmpty()
                || $href->length() === 1
                || $href->endsWith('#')
                || $href->startsWith('http://');
        };
    }

    private static function toUri(string $url): callable
    {
        return fn (Stringable $href): Uri => $href
            ->whenStartsWith('/', fn (): Stringable => $href->prepend($url))
            ->toUri();
    }

    private static function sameSite(string $domain): callable
    {
        return fn (Uri $uri): bool => str_ends_with($uri->host(), $domain);
    }

    private static function toString(): callable
    {
        return fn (Uri $uri): string => $uri->value();
    }
}
