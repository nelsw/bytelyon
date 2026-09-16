<?php

namespace Tests\Unit\Support\Sqs;

use App\Support\Sqs\Page;
use Illuminate\Support\Facades\Storage;
use Tests\TestCase;

class PageTest extends TestCase
{
    private function html(string $head, string $body): string
    {
        Storage::put('page/test', "<!DOCTYPE html><html lang=\"en\"><head>$head</head><body>$body</body></html>");
        return 'page/test';
    }

    public function test_title_url_and_domain(): void
    {
        $page = new Page(
            'https://www.example.com/articles/1',
            $this->html('<title>My Great Article - Site Name</title>', ''),
        );

        $this->assertSame('My Great Article - Site Name', $page->title());
        $this->assertSame('https://www.example.com/articles/1', $page->url);
        $this->assertSame('example.com', $page->domain());
    }

    public function test_meta_extracts_name_and_property_and_skips_invalid_tags(): void
    {
        $page = new Page('https://example.com', $this->html(
            <<<'HTML'
            <meta name="description" content="Named meta value">
            <meta property="og:title" content="Property meta value">
            <meta content="Orphan value with no name or property">
            <meta name="empty-content" content="">
            <meta charset="utf-8">
            HTML,
            '',
        ));

        $this->assertSame([
            'description' => 'Named meta value',
            'og:title' => 'Property meta value',
        ], $page->meta());
    }

    public function test_description_img_alt_and_img_src_return_first_matching_meta(): void
    {
        $page = new Page('https://example.com', $this->html(
            <<<'HTML'
            <meta name="description" content="A description here">
            <meta property="og:description" content="Should not be used">
            <meta name="twitter:image:alt" content="alt text here">
            <meta property="og:image" content="https://example.com/img.png">
            HTML,
            '',
        ));

        $this->assertSame('A description here', $page->description());
        $this->assertSame('alt text here', $page->imgAlt());
        $this->assertSame('https://example.com/img.png', $page->imgSrc());
    }

    public function test_description_img_alt_and_img_src_are_empty_when_no_meta_matches(): void
    {
        $page = new Page('https://example.com', $this->html('<title>No meta here</title>', ''));

        $this->assertSame('', $page->description());
        $this->assertSame('', $page->imgAlt());
        $this->assertSame('', $page->imgSrc());
        $this->assertSame([''], $page->keywords());
    }

    public function test_keywords_merges_trims_and_splits_all_matching_meta_values(): void
    {
        $page = new Page('https://example.com', $this->html(
            <<<'HTML'
            <meta property="article:tag" content="tag1, tag2 , tag3">
            <meta name="news_keywords" content="tag3,tag4">
            <meta name="keywords" content="tag5">
            HTML,
            '',
        ));

        $this->assertSame(
            ['tag1', 'tag2', 'tag3', 'tag3', 'tag4', 'tag5'],
            $page->keywords(),
        );
    }

    public function test_body_returns_joined_paragraph_text_from_article(): void
    {
        $page = new Page('https://example.com', $this->html('', <<<'HTML'
            <article>
                <p>First paragraph.</p>
                <p></p>
                <p>Second paragraph.</p>
            </article>
            HTML));

        $this->assertSame('First paragraph. Second paragraph.', $page->body());
    }

    public function test_body_falls_back_to_body_tag_when_no_article_or_main(): void
    {
        $page = new Page('https://example.com', $this->html('', '<p>Only body paragraph.</p>'));

        $this->assertSame('Only body paragraph.', $page->body());
    }

    public function test_links_keeps_only_same_domain_links(): void
    {
        $page = new Page('https://www.example.com/article', $this->html('', <<<'HTML'
            <a href="https://www.example.com/other-page">Internal with www</a>
            <a href="https://external.com/foo">External</a>
            <a href="">Empty href</a>
            <a href="/relative/path">Relative path</a>
            HTML));

        $this->assertSame(['https://www.example.com/other-page'], $page->links());
    }

    public function test_to_array_contains_all_expected_keys(): void
    {
        $page = new Page('https://www.example.com/article', $this->html(
            <<<'HTML'
            <title>Article Title</title>
            <meta name="description" content="Article description">
            <meta name="twitter:image:alt" content="Alt text">
            <meta property="og:image" content="https://example.com/img.png">
            <meta name="keywords" content="one,two">
            HTML,
            '<article><p>Body text.</p></article>',
        ));

        $this->assertSame([
            'body' => 'Body text.',
            'description' => 'Article description',
            'domain' => 'example.com',
            'img_alt' => 'Alt text',
            'img_url' => 'https://example.com/img.png',
            'keywords' => ['one', 'two'],
            'title' => 'Article Title',
            'url' => 'https://www.example.com/article',
        ], $page->toArray());
    }
}
