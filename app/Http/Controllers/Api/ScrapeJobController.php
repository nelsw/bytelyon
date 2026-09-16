<?php

namespace App\Http\Controllers\Api;

use App\Actions\UpdateArticle;
use App\Actions\UpdateOrCreateSitemapPage;
use App\Actions\UpdateSitemapUrls;
use App\Facades\Sqs;
use App\Http\Controllers\Controller;
use App\Http\Requests\Api\PageSaveRequest;
use App\Models\Article;
use App\Models\Bot;
use App\Models\Serp;
use App\Support\Sqs\Page as NewsPage;
use App\Support\Sqs\Page as SitemapPage;
use App\Support\Sqs\Serp as SearchPage;
use Illuminate\Http\Response;
use Illuminate\Support\Facades\Log;

class ScrapeJobController extends Controller
{

    public function serp(PageSaveRequest $request, Serp $serp): Response
    {
        $parsed = new SearchPage(
            $request->input('url'),
            $request->input('content_key'),
        );

        $serp->update([
            'data' => $parsed->data(),
            'screenshot_key' => $request->input('screenshot_key'),
            'content_key' => $parsed->contentKey,
        ]);

        if ($parsed->blocked()) {
            Log::warning('ScrapeJob: worker reported a Google block/CAPTCHA page', [
                'serp_id' => $serp->id,
                'query' => $serp->query,
            ]);
        }

        return response()->noContent();
    }

    public function news(PageSaveRequest $request, Article $article): Response
    {
        $page = new NewsPage(
            $request->input('url'),
            $request->input('content_key'),
        );

        (new UpdateArticle)(
            $article,
            $page->body(),
            $page->description(),
            $page->imgAlt(),
            $page->imgSrc(),
            $page->keywords(),
        );

        return response()->noContent();
    }

    public function sitemap(PageSaveRequest $request, Bot $bot): Response
    {
        $page = new SitemapPage(
            $request->input('url'),
            $request->input('content_key'),
        );

        (new UpdateOrCreateSitemapPage)(
            $bot,
            $page->url,
            $page->title(),
            $request->input('screenshot_key'),
            $page->meta(),
        );

        /*
         * Update the sitemap with this url if not set
         */
        if (!isset($bot->sitemap->urls[$page->url])) {
            // insert url as TRUE because it's been scraped
            $bot->sitemap->urls[$page->url] = true;
            (new UpdateSitemapUrls)($bot->sitemap);
        }

        /*
         * FIRST - check if we can fail fast
         */
        $links = $page->links();
        if ($links->isEmpty()) {
            return response()->noContent();
        }

        /*
         * SECOND - check if permitted to crawl links
         */
        if ($request->integer('depth') < 0) {
            // add these links to the sitemap as FALSE
            $links->each(fn(string $link) => $bot->sitemap->urls[$link] = false);
            // and update before returning home
            (new UpdateSitemapUrls)($bot->sitemap);
            return response()->noContent();
        }

        /*
         * LAST - transform link into a generic payload as add to end of queue
         */
        $links->transform(fn(string $link) => [
            'type' => $bot->type,
            'id' => $bot->id,
            'url' => $link,
            'depth' => $request->integer('depth') - 1,
        ])->each(Sqs::enqueueFN());

        return response()->noContent();
    }
}
