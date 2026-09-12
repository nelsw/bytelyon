<?php

namespace App\Services;

use App\Data\Html\Page;
use App\Models\Proxy;
use Exception;
use Illuminate\Container\Attributes\Singleton;
use Illuminate\Support\Facades\Http;

#[Singleton]
readonly class PageService
{
    const string USER_AGENT =
        'Mozilla/5.0 (Windows NT 10.0; Win64; x64) ' .
        'AppleWebKit/537.36 (KHTML, like Gecko) ' .
        'Chrome/124.0 ' .
        'Safari/537.36';

    public function __construct(
        private LambdaService $lambdaService,
    ) {}

    public function news(string $url, Proxy $proxy): Page
    {
        try {
            return Page::fromString($url, Http::withProxy($proxy)
                ->withUserAgent(self::USER_AGENT)
                ->get($url)
                ->throw()
                ->body());
        } catch (Exception) {
            return Page::fromArray($url, $this->lambdaService->news([$url])[0]??[]);
        }
    }
}
