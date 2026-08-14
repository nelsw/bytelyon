<?php

namespace App\Concerns;

trait SitemapValidationRules
{
    /** @return array<int, string> */
    protected function sitemapRules(): array
    {
        return [
            'bot_id' => ['sometimes', 'required', 'integer', 'exists:bots,id'],
            'domain' => ['sometimes', 'required', 'string', 'max:255'],
            'urls' => ['nullable', 'array'],
            'urls.*' => ['required', 'string', 'url'],
        ];
    }
}
