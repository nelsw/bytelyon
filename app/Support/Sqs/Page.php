<?php

namespace App\Support\Sqs;

use App\Traits\HasLinks;
use Illuminate\Support\Facades\Storage;
use Illuminate\Support\Uri;

class Page extends Html
{
    use HasLinks;

    public readonly string $url;

    public readonly string $domain;

    public function __construct(
        private readonly string $rawUrl = '',
        public readonly ?string $contentKey = null,
    ) {
        parent::__construct(Storage::get($contentKey));
        $this->url = Uri::clean($this->rawUrl);
        $this->domain = Uri::domain($this->url);
    }

    /**
     * @return array<string, string[]|string>
     */
    public function toArray(): array
    {
        return [
            'body' => $this->body(),
            'description' => $this->description(),
            'domain' => $this->domain,
            'img_alt' => $this->imgAlt(),
            'img_url' => $this->imgSrc(),
            'keywords' => $this->keywords(),
            'title' => $this->title(),
            'url' => $this->url,
        ];
    }
}
