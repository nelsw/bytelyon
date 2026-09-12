<?php

namespace App\Data\Html;

use Dom\HTMLDocument;
use Illuminate\Contracts\Support\Arrayable;
use Illuminate\Support\Facades\URL;

readonly class Page implements Arrayable
{
    public function __construct(
        public string $url,
        public string $domain,
        public string $title,
        public Meta   $meta,
        public Body   $body,
    ){}


    public static function empty(): self
    {
        return new self(
            url: '',
            domain: '',
            title: '',
            meta: Meta::empty(),
            body: Body::empty(),
        );
    }

    public static function fromArray(string $url, array $data): self
    {
        $meta = Meta::fromArray($data['meta'] ?? []);
        return new self(
            url: $url,
            domain: URL::toDomain($url),
            title: $meta->title,
            meta: $meta,
            body: Body::make($data['body'] ?? ''),
        );
    }

    public static function fromString(string $url, string $html): self
    {
        $meta = Meta::make($html);
        return new self(
            url: $url,
            domain: URL::toDomain($url),
            title: HTMLDocument::createFromString($html)->querySelector('title')?->textContent ?? $meta->title,
            meta: $meta,
            body: Body::make($html),
        );
    }

    public function isEmpty(): bool
    {
        return $this->meta->isEmpty() && $this->body->isEmpty();
    }

    public function toArray(): array
    {
        return [
            'url' => $this->url,
            'domain' => $this->domain,
            'title' => $this->title,
            'meta' => $this->meta->toArray(),
            'body' => $this->body->toArray(),
        ];
    }
}
