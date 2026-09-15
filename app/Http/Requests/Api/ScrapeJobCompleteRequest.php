<?php

namespace App\Http\Requests\Api;

use Illuminate\Contracts\Validation\ValidationRule;
use Illuminate\Foundation\Http\FormRequest;

/**
 * The one generic contract every worker (serp, news, sitemap, ...) reports
 * back through — see docker/worker/worker.py's module docstring. `depth`
 * is only meaningful for `sitemap` jobs (how many more crawl hops remain);
 * every other job type simply ignores it.
 */
class ScrapeJobCompleteRequest extends FormRequest
{
    /**
     * @return array<string, ValidationRule|array|string>
     */
    public function rules(): array
    {
        return [
            'url' => ['required', 'url', 'max:2048'],
            'screenshot_key' => ['required', 'string', 'max:1024'],
            'content_key' => ['required', 'string', 'max:1024'],
            'depth' => ['nullable', 'integer', 'min:0'],
        ];
    }
}
