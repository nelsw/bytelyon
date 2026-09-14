<?php

namespace Tests\Unit\Data\Rss;

use App\Data\Rss\BaseRssItem;
use Illuminate\Support\Carbon;
use Illuminate\Support\Facades\Log;
use Tests\TestCase;

class ConcreteRssItem extends BaseRssItem
{
    public function imageSrc(): string
    {
        return 'https://example.com/image.png';
    }

    public function publisher(): string
    {
        return 'Example Publisher';
    }

    public function source(): string
    {
        return 'example-source';
    }
}

class BaseRssItemTest extends TestCase
{
    private function makeItem(string $pubDate): ConcreteRssItem
    {
        $xml = <<<XML
        <?xml version="1.0"?>
        <item>
        <title>Item Title</title>
        <link>https://example.com/article</link>
        <description>Item description</description>
        <pubDate>{$pubDate}</pubDate>
        </item>
        XML;

        return new ConcreteRssItem($xml);
    }

    public function test_published_at_parses_valid_date(): void
    {
        $item = $this->makeItem('Mon, 01 Jan 2024 10:00:00 GMT');

        $publishedAt = $item->publishedAt();

        $this->assertInstanceOf(Carbon::class, $publishedAt);
        $this->assertSame('2024-01-01 10:00:00', $publishedAt->toDateTimeString());
    }

    public function test_published_at_logs_and_returns_raw_string_on_invalid_date(): void
    {
        Log::shouldReceive('warning')
            ->once()
            ->withArgs(fn (string $message): bool => str_starts_with($message, 'RssItem#publishedAt: '));

        $item = $this->makeItem('not a real date at all');

        $this->assertSame('not a real date at all', $item->publishedAt());
    }

    public function test_description_title_and_url(): void
    {
        $item = $this->makeItem('Mon, 01 Jan 2024 10:00:00 GMT');

        $this->assertSame('Item description', $item->description());
        $this->assertSame('Item Title', $item->title());
        $this->assertSame('https://example.com/article', $item->url());
    }

    public function test_to_array_contains_all_expected_keys(): void
    {
        $item = $this->makeItem('Mon, 01 Jan 2024 10:00:00 GMT');

        $this->assertSame([
            'description' => 'Item description',
            'img_src' => 'https://example.com/image.png',
            'published_at' => '2024-01-01 10:00:00',
            'publisher' => 'Example Publisher',
            'source' => 'example-source',
            'title' => 'Item Title',
            'url' => 'https://example.com/article',
        ], $item->toArray());
    }
}
