<?php

namespace App\Traits;

use Dom\Element;
use Illuminate\Support\Uri;

trait HasLinks
{
    /**
     * @return array<string>
     */
    public function links(): array
    {
        return collect($this->elements('a'))
            ->transform(fn (Element $e): string => trim($e->getAttribute('href') ?? ''))
            ->reject(fn (string $href): bool => empty($href))
            ->map(fn (string $href): Uri => Uri::of($href))
            ->filter(fn (Uri $uri): bool => str($uri->host())->replace('www.', '')->is($this->domain()))
            ->transform(fn (Uri $uri): string => $uri->toString())
            ->toArray();
    }
}
