<?php

namespace App\Data\Lambda;

use App\Support\Html;
use Aws\ResultInterface;
use Illuminate\Support\Facades\Storage;

readonly class Payload
{
    final private function __construct(
        public string $contentKey,
        public string $screenshotKey,
        public string $url,
        public Html $html,
    ) {}

    public static function make(ResultInterface $result): static
    {
        $output = json_decode($result->get('Payload')->getContents(), true);
        $contentKey = $output['content_key'] ?? '';
        $url = $output['url'] ?? '';
        return new static(
            $contentKey,
            $output['screenshot_key'] ?? '',
            $url,
            Html::of($url, Storage::disk('s3')->get($contentKey)),
        );
    }
}
