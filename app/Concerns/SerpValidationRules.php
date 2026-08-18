<?php

namespace App\Concerns;

trait SerpValidationRules
{
    /** @return array<int, string> */
    protected function serpRules(): array
    {
        return [
            'bot_id' => ['sometimes', 'required', 'integer', 'exists:bots,id'],
            'content_key' => ['nullable', 'string', 'max:255'],
            'data' => ['nullable', 'array'],
            'query' => ['sometimes', 'required', 'string', 'max:255'],
            'screenshot_key' => ['nullable', 'string', 'max:255'],
        ];
    }
}
