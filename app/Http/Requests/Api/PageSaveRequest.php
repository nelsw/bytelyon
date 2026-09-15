<?php

namespace App\Http\Requests\Api;

use App\Concerns\ScrapeValidationRules;
use Illuminate\Foundation\Http\FormRequest;

class PageSaveRequest extends FormRequest
{
    use ScrapeValidationRules;

    /** @return array<string, string[]> */
    public function rules(): array
    {
        return $this->scrapeRules();
    }
}
