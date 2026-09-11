<?php

namespace App\Services;

use Illuminate\Http\Client\ConnectionException;
use Illuminate\Http\Client\RequestException;
use Illuminate\Support\Facades\Http;

readonly class RssService
{
    public function items(string $class, string $url, array $query): array
    {
        try {
            $body = Http::get($url, $query)->throw()->body();
        } catch (RequestException|ConnectionException $e) {
            return [];
        }
        return (array) simplexml_load_string($body, $class)->xpath('//item');
    }
}
