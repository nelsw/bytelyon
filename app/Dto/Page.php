<?php

namespace App\Dto;

readonly class Page
{
    public function __construct(
        public string $url,
        public string $title,
        public string $body,
        public ?Meta $meta = null,
        public ?Links $links = null,
    ) {
    }
}
