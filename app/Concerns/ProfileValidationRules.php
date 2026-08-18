<?php

namespace App\Concerns;

use App\Models\User;
use Illuminate\Contracts\Validation\ValidationRule;
use Illuminate\Validation\Rule;

trait ProfileValidationRules
{
    /** @return array<string, array<int, array>> */
    protected function profileRules(?int $userId = null): array
    {
        return [
            'name' => $this->nameRules(),
            'email' => $this->emailRules($userId),
            'img_url' => $this->imgUrlRules(),
        ];
    }

    /** @return array<int, string> */
    protected function imgUrlRules(): array
    {
        return ['nullable', 'string', 'url', 'max:1024'];
    }

    /** @return array<int, string> */
    protected function nameRules(): array
    {
        return ['required', 'string', 'max:255'];
    }

    /** @return array<int, ValidationRule|string> */
    protected function emailRules(?int $userId = null): array
    {
        return [
            'required',
            'string',
            'email',
            'max:255',
            $userId === null
                ? Rule::unique(User::class)
                : Rule::unique(User::class)->ignore($userId),
        ];
    }
}
