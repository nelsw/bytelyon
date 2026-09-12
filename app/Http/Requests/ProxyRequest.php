<?php

namespace App\Http\Requests;

use Illuminate\Foundation\Http\FormRequest;
use Illuminate\Validation\Rule;

class ProxyRequest extends FormRequest
{
    public function rules(): array
    {
        return [
            'protocol' => ['required', Rule::in(['http', 'https', 'socks5'])],
            'server' => ['required'],
            'port' => ['nullable', 'integer', 'between:1,65535'],
            'username' => ['nullable'],
            'password' => ['nullable'],
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
