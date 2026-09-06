<?php

namespace App\Services\Bots\News\Rss;

use App\Enums\NewsSource;
use App\Models\Article;
use Carbon\Exceptions\InvalidDateException;
use Dom\Element;
use Dom\HTMLDocument;
use DOMDocument;
use DOMElement;
use Illuminate\Container\Attributes\Singleton;
use Illuminate\Http\Client\ConnectionException;
use Illuminate\Http\Client\PendingRequest;
use Illuminate\Http\Client\RequestException;
use Illuminate\Support\Carbon;
use Illuminate\Support\Facades\Cache;
use Illuminate\Support\Facades\Http;
use Illuminate\Support\Facades\Log;
use Illuminate\Support\Str;
use JsonException;
use RuntimeException;
use Throwable;

#[Singleton]
readonly class GoogleRssService
{
    private const string RSS_URL = "https://news.google.com/rss/search";
    private const string ENDPOINT = 'https://news.google.com/_/DotsSplashUi/data/batchexecute';
    private const string LINK_REGEX = '~/articles/(?P<encoded_url>[^?]+)~';
    private const string USER_AGENT = 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 '
    . '(KHTML, like Gecko) Chrome/124.0 Safari/537.36';
    private const int RETRIES = 2;
    private const int TIMEOUT = 30;

    /**
     * @param string $query
     * @param Carbon $last
     * @param array $blacklist
     *
     * @return array<string, Article>
     *
     * @throws ConnectionException
     * @throws RequestException
     */
    public function fetch(string $query, Carbon $last, array $blacklist): array
    {
        $body = Http::get(url: self::RSS_URL, query: [
            'q' => $query,
            'hl' => 'en-US',
            'gl' => 'US',
            'ceid' => 'US:en',
        ])->throw()->body();

        $arr = [];
        $xml = simplexml_load_string($body);
        for ($i = 0; $i < count($xml->channel->item); $i++) {

            $item = $xml->channel->item[$i];

            $publisher = str((string)$item->description)
                ->trim()
                ->after('font')
                ->after('>')
                ->before('<')
                ->replace("\n", ' ');

            while ($publisher->contains("  ")) {
                $publisher = $publisher->replace("  ", " ");
            }
            $publisher = $publisher->trim()->toString();

            $title = str((string)$item->title)
                ->trim()
                ->replace("\n", ' ');

            while ($title->contains('  ')) {
                $title = $title->replace('  ', ' ');
            }
            $title = $title
                ->remove(" - $publisher")
                ->trim()
                ->toString();

            try {
                $date = Carbon::parse((string)$item->pubDate);
            } catch (InvalidDateException) {
                continue;
            }

            if ($date->isBefore($last)) {
                continue;
            }

            foreach ($blacklist as $keyword) {
                if (str_contains($title, $keyword)) {
                    continue 2;
                }
            }

            $arr[] = [
                "pubDate" => $date,
                "title" => $title,
                "link" => $this->decoded((string)$item->link),
                "source" => NewsSource::GoogleNews->value,
                "publisher" => $publisher,
            ];
        }
        return $arr;
    }

    /**
     * Cached variant. Only successful decodes are stored, so a transient failure
     * doesn't poison the key.
     */
    private function decoded(string $link): string
    {
        $key = 'gnews:decoded:' . sha1($link);
        if (is_string($cached = Cache::get($key))) {
            return $cached;
        }

        $url = $this->decode($link);
        if (!empty($url)) {
            Cache::put($key, $url, now()->addDay());
        }
        return $url;
    }

    private function decode(string $link): string|null
    {
        $context = ['gURL' => $link];

        if (preg_match(self::LINK_REGEX, $link, $matches) !== 1) {
            Log::warning('failed to match gstatic regex', $context);

            return '';
        }

        try {
            $node = HTMLDocument::createFromString(
                source: $this->client()->get($link)->body(),
                options: LIBXML_NOERROR,
            );
        } catch (Throwable $e) {
            Log::warning('failed to parse gstatic html', [...$context, 'error' => $e->getMessage()]);

            return '';
        }

        $encoded = $matches['encoded_url'];
        try {
            $url = $this->decodeNode($node, $encoded);
        } catch (Throwable $e) {
            Log::warning('failed to decode gstatic node', [...$context, 'error' => $e->getMessage()]);

            return '';
        }

        Log::debug('decoded gstatic url', ['url' => $url]);
        return rtrim($url, '/');
    }

    /**
     * Equivalent of decodeNode: walk every <c-wiz> in document order and try each one,
     * returning the first success. The Go version keeps traversing siblings when a node
     * fails to decode, so a failure here is not fatal until every node is exhausted.
     */
    private function decodeNode(DOMDocument|HTMLDocument $node, string $encodedText): string|null
    {
        /**
         * Reads data-n-a-sg / data-n-a-ts off the first child element of the <c-wiz>.
         *
         * Go's n.FirstChild can land on a text node (yielding empty values); firstElementChild
         * skips whitespace, which is what the original was reaching for anyway.
         *
         * @return array{0: string, 1: string} [signature, timestamp]
         */
        $ƒ = function(DOMElement|Element $wiz): array {
            return [
                $wiz->firstElementChild?->getAttribute('data-n-a-sg') ?? '',
                $wiz->firstElementChild?->getAttribute('data-n-a-ts') ?? '',
            ];
        };

        foreach ($node->getElementsByTagName('c-wiz') as $wiz) {
            [$sg, $ts] = $ƒ($wiz);
            try {
                return $this->decodeParts($sg, $ts, $encodedText);
            } catch (Throwable) {}
        }
        return null;
    }


    /**
     * Equivalent of decodeParts: hits the batch-execute endpoint and unwraps the
     * doubly encoded JSON response.
     * @throws JsonException
     * @throws ConnectionException
     */
    protected function decodeParts(string $signature, string $timestamp, string $base64Str): string|null
    {
        if ($signature === '' || $timestamp === '') {
            // Interpolated raw into the payload below; an empty timestamp yields invalid JSON.
            throw new RuntimeException('missing signature or timestamp');
        }

        $request = sprintf(
            '["garturlreq",[["X","X",["X","X"],null,null,1,1,"US:en",null,1,null,null,null,null,null,0,1],'
            . '"X","X",1,[1,1,1],1,1,null,0,0,null,0],"%s",%s,"%s"]',
            $base64Str,
            $timestamp,
            $signature,
        );

        // JSON_UNESCAPED_SLASHES keeps the payload byte-identical to Go's json.Marshal.
        $body = json_encode([[['Fbv4je', $request]]], JSON_UNESCAPED_SLASHES | JSON_THROW_ON_ERROR);

        $raw = $this->client()
            ->asForm()
            ->post(self::ENDPOINT, ['f.req' => $body])
            ->body();

        $line = Str::of($raw)->explode("\n\n")->get(1);
        if (!is_string($line)) {
            throw new RuntimeException('unexpected batch-execute response format');
        }

        $payload = json_decode($line, true, 512, JSON_THROW_ON_ERROR);
        if (!is_array($payload) || $payload === []) {
            throw new RuntimeException('empty payload');
        }

        $innerJson = data_get($payload, '0.2');
        if (!is_string($innerJson)) {
            throw new RuntimeException('missing inner json string');
        }

        $url = data_get(json_decode($innerJson, true, 512, JSON_THROW_ON_ERROR), 1);
        if (!is_string($url) || $url === '') {
            throw new RuntimeException('decoded url not string');
        }

        return $url;
    }

    /**
     * Shared client config. ->throw() turns 4xx/5xx into RequestException, which the
     * callers in decode() already handle as a Throwable.
     */
    private function client(): PendingRequest
    {
        return Http::withUserAgent(self::USER_AGENT)
            ->connectTimeout(5)
            ->timeout(self::TIMEOUT)
            ->retry(self::RETRIES, 250, throw: false)
            ->throw();
    }
}
