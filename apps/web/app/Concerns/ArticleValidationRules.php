<?php

namespace App\Concerns;

use Carbon\CarbonInterface;
use Illuminate\Validation\Rule;

trait ArticleValidationRules
{
    /** @return array<int, string> */
    protected function articleRules(?CarbonInterface $after = null): array
    {
        return [
            'body' => ['nullable', 'string'],
            'bot_id' => ['sometimes', 'required', 'integer', 'exists:bots,id'],
            'description' => ['nullable', 'string'],
            'img_alt' => ['nullable', 'string', 'max:1024'],
            'img_url' => ['nullable', 'string', 'max:2048'],
            'keywords' => ['nullable', 'array'],
            'url' => ['required', 'string', 'url'],
            'published_at' => ['required', 'date', Rule::date()->afterOrEqual($after ?? now()->subCentury())],
            'publisher' => ['nullable', 'string', 'max:255'],
            'source' => ['sometimes', 'required', 'string', 'max:255'],
            'title' => ['required', 'string', 'max:255'],
        ];
    }
}
