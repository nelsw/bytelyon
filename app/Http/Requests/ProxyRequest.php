<?php

namespace App\Http\Requests;

use Illuminate\Foundation\Http\FormRequest;

class ProxyRequest extends FormRequest
{
    public function rules(): array
    {
        return [
            'server' => ['required'],
            'username' => ['required'],
            'password' => ['required'],
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
