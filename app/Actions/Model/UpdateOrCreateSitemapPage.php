<?php

namespace App\Actions\Model;

use App\Concerns\SitemapPageValidationRules;
use App\Models\Bot;
use App\Models\Page;
use App\Models\Sitemap;
use Illuminate\Support\Facades\Log;
use Illuminate\Support\Facades\Validator;

class UpdateOrCreateSitemapPage
{
    use SitemapPageValidationRules;

    /**
     * @param  string[]  $meta
     */
    public function __invoke(
        Bot $bot,
        string $url,
        string $title,
        string $screenshotKey,
        array $meta,
    ): void {
        $values = Validator::validate([
            'domain' => $bot->sitemap?->domain,
            'meta' => $meta,
            'pageable_id' => $bot->sitemap?->id,
            'pageable_type' => Sitemap::class,
            'screenshot_key' => $screenshotKey,
            'title' => $title,
            'url' => $url,
        ], $this->sitemapPageRules());

        /** @var Page $page */
        $page = $bot->sitemap?->pages()->updateOrCreate([
            'url' => $url,
            'pageable_type' => Sitemap::class,
            'pageable_id' => $bot->sitemap?->id,
        ], $values);

        Log::debug('Page '.($page->wasRecentlyCreated ? 'created' : 'updated'), [
            'id' => $page->id,
            'sitemap' => $bot->sitemap?->id,
            'domain' => $bot->sitemap?->domain,
            'title' => $page->title,
            'url' => $page->url,
        ]);
    }
}
