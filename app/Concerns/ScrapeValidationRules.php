<?php

namespace App\Concerns;

trait ScrapeValidationRules
{
    /** @return array<string, string[]> */
    protected function scrapeRules(): array
    {
        return [
            'screenshot_key' => ['required', 'string', 'max:1024'],
            'content_key' => ['required', 'string', 'max:1024'],
            'url' => ['required', 'url', 'max:2048'],
            'depth' => ['nullable', 'integer', 'min:0'],
        ];
    }
}
