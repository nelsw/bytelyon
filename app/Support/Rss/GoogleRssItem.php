<?php

namespace App\Support\Rss;

use Dom\Element;
use Dom\HTMLDocument;
use Illuminate\Http\Client\ConnectionException;
use Illuminate\Http\Client\PendingRequest;
use Illuminate\Support\Facades\Cache;
use Illuminate\Support\Facades\Http;
use Illuminate\Support\Facades\Log;
use Illuminate\Support\Str;
use JsonException;
use RuntimeException;
use Throwable;

class GoogleRssItem extends RssItem
{
    private const int LIBXML_NOERROR = 32;

    private const string ENDPOINT = 'https://news.google.com/_/DotsSplashUi/data/batchexecute';

    private const string LINK_REGEX = '~/articles/(?P<encoded_url>[^?]+)~';

    private const string USER_AGENT = 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 '
        .'(KHTML, like Gecko) Chrome/124.0 Safari/537.36';

    private const int RETRIES = 2;

    private const int TIMEOUT = 30;

    public function imageSrc(): string
    {
        return '';
    }

    public function publisher(): string
    {
        return (string) $this->source;
    }

    public function title(): string
    {
        return rescue(fn () => explode(' - ', $this->title)[0], (string) $this->title);
    }

    public function url(): string
    {
        return $this->decoded((string) $this->link);
    }

    private function decoded(string $link): string
    {
        $key = 'gnews:decoded:'.sha1($link);
        if (is_string($cached = Cache::get($key))) {
            return $cached;
        }

        $url = $this->decode($link);
        if (! empty($url)) {
            Cache::put($key, $url, now()->addDay());
        }
        return $url;
    }

    private function decode(string $link): ?string
    {
        $context = ['gURL' => $link];

        if (preg_match(self::LINK_REGEX, $link, $matches) !== 1) {
            Log::warning('failed to match gstatic regex', $context);

            return '';
        }

        try {
            $node = HTMLDocument::createFromString(
                source: $this->client()->get($link)->body(),
                options: self::LIBXML_NOERROR,
            );
        } catch (Throwable $e) {
            Log::warning('failed to parse gstatic html', [...$context, 'error' => $e->getMessage()]);

            return '';
        }

        $encoded = $matches['encoded_url'];
        $url = '';

        try {
            $url = $this->decodeNode($node, $encoded);
            // @codeCoverageIgnoreStart
        } catch (Throwable $e) {
            Log::warning('failed to decode gstatic node', [...$context, 'error' => $e->getMessage()]);

            return '';
        }
        // @codeCoverageIgnoreEnd

        Log::debug('decoded gstatic url', ['url' => $url]);
        return rtrim($url, '/');
    }

    private function decodeNode(HTMLDocument $node, string $encodedText): ?string
    {
        /**
         * @return array{0: string, 1: string} [signature, timestamp]
         */
        $ƒ = function (Element $wiz): array {
            return [
                $wiz->firstElementChild?->getAttribute('data-n-a-sg') ?? '',
                $wiz->firstElementChild?->getAttribute('data-n-a-ts') ?? '',
            ];
        };

        foreach ($node->getElementsByTagName('c-wiz') as $wiz) {
            [$sg, $ts] = $ƒ($wiz);

            try {
                return $this->decodeParts($sg, $ts, $encodedText);
            } catch (Throwable) {
            }
        }
        return null;
    }

    /**
     * @throws JsonException
     * @throws ConnectionException
     */
    protected function decodeParts(string $signature, string $timestamp, string $base64Str): ?string
    {
        if ($signature === '' || $timestamp === '') {
            throw new RuntimeException('missing signature or timestamp');
        }

        $request = sprintf(
            '["garturlreq",[["X","X",["X","X"],null,null,1,1,"US:en",null,1,null,null,null,null,null,0,1],'
            .'"X","X",1,[1,1,1],1,1,null,0,0,null,0],"%s",%s,"%s"]',
            $base64Str,
            $timestamp,
            $signature,
        );

        $body = json_encode([[['Fbv4je', $request]]], JSON_UNESCAPED_SLASHES | JSON_THROW_ON_ERROR);

        $raw = $this->client()
            ->asForm()
            ->post(self::ENDPOINT, ['f.req' => $body])
            ->body();

        $line = Str::of($raw)->explode("\n\n")->get(1);
        if (! is_string($line)) {
            throw new RuntimeException('unexpected batch-execute response format');
        }

        $payload = json_decode($line, true, 512, JSON_THROW_ON_ERROR);
        if (! is_array($payload) || $payload === []) {
            throw new RuntimeException('empty payload');
        }

        $innerJson = data_get($payload, '0.2');
        if (! is_string($innerJson)) {
            throw new RuntimeException('missing inner json string');
        }

        $url = data_get(json_decode($innerJson, true, 512, JSON_THROW_ON_ERROR), 1);
        if (! is_string($url) || $url === '') {
            throw new RuntimeException('decoded url not string');
        }

        return $url;
    }

    private function client(): PendingRequest
    {
        return Http::withUserAgent(self::USER_AGENT)
            ->connectTimeout(5)
            ->timeout(self::TIMEOUT)
            ->retry(self::RETRIES, 250, throw: false)
            ->throw();
    }
}
