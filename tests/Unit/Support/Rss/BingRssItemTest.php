<?php

namespace Tests\Unit\Support\Rss;

use App\Enums\NewsSource;
use App\Support\Rss\BingRssItem;
use Tests\TestCase;

class BingRssItemTest extends TestCase
{
    private function makeItem(string $link, string $sourceTag = ''): BingRssItem
    {
        return new BingRssItem(<<<XML
<?xml version="1.0"?>
<item xmlns:News="http://schemas.microsoft.com/HealthVault/2007/thing/News">
    <title>Bing Article Title</title>
    <link>$link</link>
    <description>Bing description</description>
    <pubDate>Mon, 01 Jan 2024 10:00:00 GMT</pubDate>
    <News:Image>SomeImage</News:Image>
    {$sourceTag}
</item>
XML);
    }

    public function test_publisher_and_image_src_return_first_news_source_node(): void
    {
        $item = $this->makeItem(
            link: 'https://www.bing.com/news/apiclick.aspx?ref=X&amp;url=https%3A%2F%2Fexample.com%2Farticle%2F&amp;cc=us',
            sourceTag: '<News:Source>Example Publisher</News:Source>',
        );

        $this->assertSame('Example Publisher', $item->publisher());
        $this->assertSame('SomeImage', $item->imageSrc());
    }

    public function test_publisher_and_image_src_default_to_empty_string_when_missing(): void
    {
        $item = $this->makeItem(
            link: 'https://www.bing.com/news/apiclick.aspx?ref=X&amp;cc=us',
        );

        $this->assertSame('', $item->publisher());
        $this->assertSame('SomeImage', $item->imageSrc());
    }

    public function test_source_returns_bing_news_enum_value(): void
    {
        $item = $this->makeItem(link: 'https://www.bing.com/news/apiclick.aspx?ref=X&amp;cc=us');

        $this->assertSame(NewsSource::BingNews->value, $item->source());
    }

    public function test_url_decodes_url_query_parameter(): void
    {
        $item = $this->makeItem(
            link: 'https://www.bing.com/news/apiclick.aspx?ref=X&amp;url=https%3A%2F%2Fexample.com%2Farticle%2F&amp;cc=us',
        );

        $this->assertSame('https://example.com/article', $item->url());
    }

    public function test_url_falls_back_to_raw_link_when_no_url_query_parameter(): void
    {
        $link = 'https://www.bing.com/news/apiclick.aspx?ref=X&amp;cc=us';

        $item = $this->makeItem(link: $link);

        $this->assertSame('https://www.bing.com/news/apiclick.aspx?ref=X&cc=us', $item->url());
    }
}
