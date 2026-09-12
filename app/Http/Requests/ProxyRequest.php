<?php

namespace App\Http\Requests;

use Illuminate\Foundation\Http\FormRequest;
use Illuminate\Validation\Rule;

class ProxyRequest extends FormRequest
{
    public function rules(): array
    {
        return [
            'scheme' => ['required', Rule::in(['http', 'https', 'socks5'])],
            'host' => ['required'],
            'port' => ['nullable', 'integer', 'between:1,65535'],
            'user' => ['nullable'],
            'pass' => ['nullable'],
            'bypass' => ['nullable'],
            'user_id' => ['required', 'exists:users'],
            'name' => ['required'],
        ];
    }

    public function authorize(): bool
    {
        return true;
    }
}
