<?php

namespace App\Concerns;

use App\Enums\BotType;
use App\Enums\FrequencyType;
use Illuminate\Contracts\Validation\ValidationRule;
use Illuminate\Validation\Rule;

trait BotValidationRules
{
    /** @return array<int, ValidationRule|string> */
    protected function createRules(): array
    {
        return [
            ...$this->updateRules(),
            ...[
                'type' => ['required', Rule::enum(BotType::class)],
                'query' => [
                    'required',
                    'string',
                    'min:7',
                    'max:255',
                    Rule::unique('bots')
                        ->where(fn ($query) => $query
                            ->where('user_id', request()->user()?->id)
                            ->where('type', request()->input('type'))),
                ],
            ],
        ];
    }

    /** @return array<int, ValidationRule|string> */
    protected function updateRules(): array
    {
        return [
            'blacklist' => ['nullable', 'string'],
            'enabled' => ['required', 'boolean'],
            'headless' => ['required', 'boolean'],
            'frequency' => ['required', Rule::enum(FrequencyType::class)],
        ];
    }
}
