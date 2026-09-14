<?php

namespace Tests\Unit\Data\Rss;

use App\Data\Rss\GoogleRssItem;
use App\Enums\NewsSource;
use Illuminate\Http\Client\ConnectionException;
use Illuminate\Support\Facades\Cache;
use Illuminate\Support\Facades\Http;
use Tests\TestCase;

class DecodePartsExposingGoogleRssItem extends GoogleRssItem
{
    public function callDecodeParts(string $signature, string $timestamp, string $base64Str): ?string
    {
        return $this->decodeParts($signature, $timestamp, $base64Str);
    }
}

class GoogleRssItemTest extends TestCase
{
    private const string ENDPOINT = 'https://news.google.com/_/DotsSplashUi/data/batchexecute';

    protected function tearDown(): void
    {
        Http::fake();

        parent::tearDown();
    }

    private function makeXml(string $link, string $title = 'Article Title - Google News', string $source = 'Example Source'): string
    {
        return <<<XML
        <?xml version="1.0"?>
        <item>
        <title>{$title}</title>
        <link>{$link}</link>
        <description>Google description</description>
        <pubDate>Mon, 01 Jan 2024 10:00:00 GMT</pubDate>
        <source>{$source}</source>
        </item>
        XML;
    }

    private function batchExecuteBody(?string $decodedUrl): string
    {
        $innerJson = json_encode([0, $decodedUrl]);
        $payload = [['wrb.fr', 'Fbv4je', $innerJson, null, null, null, 'generic']];
        $line = json_encode($payload);

        return ")]}'\n\n{$line}";
    }

    public function test_image_src_is_always_empty(): void
    {
        $item = new GoogleRssItem($this->makeXml('https://news.google.com/rss/articles/ENC1?hl=en'));

        $this->assertSame('', $item->imageSrc());
    }

    public function test_publisher_returns_source_element(): void
    {
        $item = new GoogleRssItem($this->makeXml('https://news.google.com/rss/articles/ENC1?hl=en', source: 'My Source'));

        $this->assertSame('My Source', $item->publisher());
    }

    public function test_title_strips_suffix_after_dash(): void
    {
        $item = new GoogleRssItem($this->makeXml('https://news.google.com/rss/articles/ENC1?hl=en', title: 'Real Title - Google News'));

        $this->assertSame('Real Title', $item->title());
    }

    public function test_source_returns_google_news_enum_value(): void
    {
        $item = new GoogleRssItem($this->makeXml('https://news.google.com/rss/articles/ENC1?hl=en'));

        $this->assertSame(NewsSource::GoogleNews->value, $item->source());
    }

    public function test_url_returns_empty_string_when_link_has_no_articles_segment(): void
    {
        $link = 'https://news.google.com/rss/topics/SOMETOPIC?hl=en';
        Cache::forget('gnews:decoded:'.sha1($link));

        $item = new GoogleRssItem($this->makeXml($link));

        $this->assertSame('', $item->url());
        Http::assertNothingSent();
    }

    public function test_url_returns_empty_string_when_http_request_fails(): void
    {
        $link = 'https://news.google.com/rss/articles/ENCFAIL?hl=en';
        Cache::forget('gnews:decoded:'.sha1($link));

        Http::fake([
            'https://news.google.com/rss/articles/*' => function () {
                throw new ConnectionException('connection failed');
            },
        ]);

        $item = new GoogleRssItem($this->makeXml($link));

        $this->assertSame('', $item->url());
    }

    public function test_url_returns_empty_string_when_page_has_no_c_wiz_element(): void
    {
        $link = 'https://news.google.com/rss/articles/ENCNOWIZ?hl=en';
        Cache::forget('gnews:decoded:'.sha1($link));

        Http::fake([
            'https://news.google.com/rss/articles/*' => Http::response(
                '<!DOCTYPE html><html><body><p>no wiz here</p></body></html>',
                200,
            ),
        ]);

        $item = new GoogleRssItem($this->makeXml($link));

        $this->assertSame('', $item->url());
    }

    public function test_url_returns_empty_string_when_c_wiz_is_missing_signature_data(): void
    {
        $link = 'https://news.google.com/rss/articles/ENCEMPTYWIZ?hl=en';
        Cache::forget('gnews:decoded:'.sha1($link));

        Http::fake([
            'https://news.google.com/rss/articles/*' => Http::response(
                '<!DOCTYPE html><html><body><c-wiz><div></div></c-wiz></body></html>',
                200,
            ),
        ]);

        $item = new GoogleRssItem($this->makeXml($link));

        $this->assertSame('', $item->url());
    }

    public function test_url_decodes_and_caches_successfully(): void
    {
        $link = 'https://news.google.com/rss/articles/ENCSUCCESS?hl=en';
        $cacheKey = 'gnews:decoded:'.sha1($link);
        Cache::forget($cacheKey);

        Http::fake([
            'https://news.google.com/rss/articles/*' => Http::response(
                '<!DOCTYPE html><html><body><c-wiz><div data-n-a-sg="SIG" data-n-a-ts="12345"></div></c-wiz></body></html>',
                200,
            ),
            self::ENDPOINT => Http::response($this->batchExecuteBody('https://example.com/decoded-article/'), 200),
        ]);

        $item = new GoogleRssItem($this->makeXml($link));

        $this->assertSame('https://example.com/decoded-article', $item->url());
        $this->assertSame('https://example.com/decoded-article', Cache::get($cacheKey));

        Http::assertSentCount(2);

        $this->assertSame('https://example.com/decoded-article', $item->url());
        Http::assertSentCount(2);

        Cache::forget($cacheKey);
    }

    public function test_url_returns_cached_value_without_any_http_calls(): void
    {
        $link = 'https://news.google.com/rss/articles/ENCCACHED?hl=en';
        $cacheKey = 'gnews:decoded:'.sha1($link);
        Cache::put($cacheKey, 'https://cached.example.com/article', now()->addDay());

        Http::fake();

        $item = new GoogleRssItem($this->makeXml($link));

        $this->assertSame('https://cached.example.com/article', $item->url());
        Http::assertNothingSent();

        Cache::forget($cacheKey);
    }

    public function test_decode_parts_throws_when_signature_or_timestamp_missing(): void
    {
        $item = new DecodePartsExposingGoogleRssItem($this->makeXml('https://news.google.com/rss/articles/ENC1?hl=en'));

        $this->expectException(\RuntimeException::class);
        $this->expectExceptionMessage('missing signature or timestamp');

        $item->callDecodeParts('', '', 'base64');
    }

    public function test_decode_parts_throws_when_batch_execute_response_has_no_second_line(): void
    {
        Http::fake([self::ENDPOINT => Http::response('single-line-only', 200)]);

        $item = new DecodePartsExposingGoogleRssItem($this->makeXml('https://news.google.com/rss/articles/ENC1?hl=en'));

        $this->expectException(\RuntimeException::class);
        $this->expectExceptionMessage('unexpected batch-execute response format');

        $item->callDecodeParts('sig', 'ts', 'base64');
    }

    public function test_decode_parts_throws_when_payload_is_empty(): void
    {
        Http::fake([self::ENDPOINT => Http::response("line0\n\n[]", 200)]);

        $item = new DecodePartsExposingGoogleRssItem($this->makeXml('https://news.google.com/rss/articles/ENC1?hl=en'));

        $this->expectException(\RuntimeException::class);
        $this->expectExceptionMessage('empty payload');

        $item->callDecodeParts('sig', 'ts', 'base64');
    }

    public function test_decode_parts_throws_when_inner_json_missing(): void
    {
        $line = json_encode([['wrb.fr', 'Fbv4je', null]]);
        Http::fake([self::ENDPOINT => Http::response("line0\n\n{$line}", 200)]);

        $item = new DecodePartsExposingGoogleRssItem($this->makeXml('https://news.google.com/rss/articles/ENC1?hl=en'));

        $this->expectException(\RuntimeException::class);
        $this->expectExceptionMessage('missing inner json string');

        $item->callDecodeParts('sig', 'ts', 'base64');
    }

    public function test_decode_parts_throws_when_decoded_url_is_not_a_string(): void
    {
        Http::fake([self::ENDPOINT => Http::response($this->batchExecuteBody(null), 200)]);

        $item = new DecodePartsExposingGoogleRssItem($this->makeXml('https://news.google.com/rss/articles/ENC1?hl=en'));

        $this->expectException(\RuntimeException::class);
        $this->expectExceptionMessage('decoded url not string');

        $item->callDecodeParts('sig', 'ts', 'base64');
    }

    public function test_decode_parts_returns_decoded_url_on_success(): void
    {
        Http::fake([self::ENDPOINT => Http::response($this->batchExecuteBody('https://example.com/direct-decode/'), 200)]);

        $item = new DecodePartsExposingGoogleRssItem($this->makeXml('https://news.google.com/rss/articles/ENC1?hl=en'));

        $this->assertSame('https://example.com/direct-decode/', $item->callDecodeParts('sig', 'ts', 'base64'));
    }
}
