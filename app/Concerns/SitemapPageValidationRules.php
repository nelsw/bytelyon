<?php

namespace App\Concerns;

use App\Models\Sitemap;
use Illuminate\Validation\Rule;

trait SitemapPageValidationRules
{
    /** @return array<int, string> */
    protected function sitemapPageRules(): array
    {
        return [
            'domain' => ['nullable', 'string', 'max:255'],
            'meta' => ['nullable', 'array'],
            'pageable_id' => ['sometimes', 'required', 'integer', 'exists:sitemaps,id'],
            'pageable_type' => ['sometimes', 'required', 'string', Rule::in(Sitemap::class)],
            'screenshot_key' => ['nullable', 'string', 'ends_with:png'],
            'title' => ['required', 'string', 'max:1025'],
            'url' => ['required', 'string', 'url'],
        ];
    }
}
