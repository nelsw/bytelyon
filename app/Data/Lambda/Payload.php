<?php

namespace App\Data\Lambda;

use Aws\ResultInterface;

readonly class Payload
{
    final private function __construct(
        public string $contentKey,
        public string $screenshotKey,
        public string $url,
    ) {}

    public static function make(ResultInterface $result): static
    {
        $output = json_decode($result->get('Payload')->getContents(), true);
        return new static(
            $output['content_key'] ?? '',
            $output['screenshot_key'] ?? '',
            $output['url'] ?? '',
        );
    }
}
