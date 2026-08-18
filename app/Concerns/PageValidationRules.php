<?php

namespace App\Concerns;

trait PageValidationRules
{
    /** @return array<int, string> */
    protected function pageRules(): array
    {
        return [
            'domain' => ['nullable', 'string', 'max:255'],
            'meta' => ['nullable', 'array'],
            'screenshot_key' => ['nullable', 'string', 'max:255'],
            'title' => ['required', 'string', 'max:1025'],
            'url' => ['required', 'string', 'url'],
            'index' => ['nullable', 'integer'],
            'kind' => ['nullable', 'string', 'max:255'],
        ];
    }
}
