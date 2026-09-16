<?php

namespace App\Support\Sqs;

use App\Traits\HasLinks;
use App\Traits\HasUrl;
use Illuminate\Support\Facades\Storage;

class Page extends Html
{
    use HasLinks, HasUrl;

    public function __construct(
        public readonly string $url = '',
        public readonly ?string $contentKey = null,
    ) {
        parent::__construct(Storage::get($contentKey));
    }

    /**
     * @return array<string, string[]|string>
     */
    public function toArray(): array
    {
        return [
            'body' => $this->body(),
            'description' => $this->description(),
            'domain' => $this->domain(),
            'img_alt' => $this->imgAlt(),
            'img_url' => $this->imgSrc(),
            'keywords' => $this->keywords(),
            'title' => $this->title(),
            'url' => $this->url,
        ];
    }
}
