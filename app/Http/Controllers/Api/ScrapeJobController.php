<?php

namespace App\Http\Controllers\Api;

use App\Actions\UpdateOrCreateSitemapPage;
use App\Actions\UpdateSitemap;
use App\Concerns\ScrapeValidationRules;
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
    use ScrapeValidationRules;

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

        $article->update($page->toArray());

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

        /** @var array<string, bool> $known */
        $known = $bot->sitemap?->pages()->pluck('url')->mapWithKeys(fn (string $u) => [$u => true])->all() ?? [];

        (new UpdateSitemap)($bot, $known);

        $depth = $request->integer('depth', 0);
        if ($depth > 0) {
            foreach (array_keys($page->links()) as $link) {
                if (isset($known[$link])) {
                    continue;
                }
                Sqs::enqueueScrape('sitemap', $bot->id, ['url' => $link, 'depth' => $depth - 1]);
            }
        }

        return response()->noContent();
    }
}
