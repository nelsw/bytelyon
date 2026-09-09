<?php

namespace App\Services;

use Closure;
use Illuminate\Container\Attributes\Singleton;
use Illuminate\Http\Client\ConnectionException;
use Illuminate\Http\Client\RequestException;
use Illuminate\Support\Facades\Http;
use SimpleXMLElement;

#[Singleton]
readonly class XmlService
{
    public function fetch(
        string $url,
        array $query = [],
    ): false|SimpleXMLElement {

        try {
            $body = Http::get($url, $query)->throw()->body();
        } catch (RequestException|ConnectionException $e) {
            return false;
        }

        return simplexml_load_string($body);
    }
}
