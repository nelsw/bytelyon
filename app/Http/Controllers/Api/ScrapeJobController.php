<?php

namespace App\Http\Controllers\Api;

use App\Actions\UpdateOrCreateSitemapPage;
use App\Actions\UpdateSitemap;
use App\Data\Html\Page as ArticlePage;
use App\Facades\Sqs;
use App\Http\Controllers\Controller;
use App\Http\Requests\Api\ScrapeJobCompleteRequest;
use App\Models\Article;
use App\Models\Bot;
use App\Models\Serp;
use App\Support\Html;
use App\Support\Serp as SerpParser;
use Illuminate\Http\Response;
use Illuminate\Support\Facades\Log;
use Illuminate\Support\Facades\Storage;

/**
 * Callback endpoints every local worker (see docker/worker/worker.py)
 * reports back through once it finishes a job pulled from the shared
 * bytelyon-scrape-jobs SQS queue. Authenticated by a `worker`-ability
 * Sanctum token rather than resource ownership — the queue is shared/global
 * across all users' bots, so any authorized worker may report a result for
 * any record, matching the flat trust model already established by the
 * `abilities:worker` route group (see routes/api.php).
 *
 * Every worker, regardless of job type, reports the exact same minimal
 * shape (url, screenshot_key, content_key) — no parsing happens in Python
 * at all anymore. All of it lives here instead: App\Support\Serp for SERP
 * structure, App\Data\Html\Page for article content, App\Support\Html for
 * sitemap link discovery. One language, one place, easy to test and
 * iterate on without rebuilding/redeploying a Python worker image.
 */
class ScrapeJobController extends Controller
{
    public function serp(ScrapeJobCompleteRequest $request, Serp $serp): Response
    {
        $contentKey = $request->string('content_key')->value();
        $html = Storage::disk('s3')->get($contentKey);
        $parsed = SerpParser::of($request->string('url')->value(), $html);

        $serp->update([
            'data' => $parsed->data(),
            'screenshot_key' => $request->string('screenshot_key')->value(),
            'content_key' => $contentKey,
        ]);

        if ($parsed->blocked()) {
            Log::warning('ScrapeJob: worker reported a Google block/CAPTCHA page', [
                'serp_id' => $serp->id,
                'query' => $serp->query,
            ]);
        }

        return response()->noContent();
    }

    public function news(ScrapeJobCompleteRequest $request, Article $article): Response
    {
        $url = $request->string('url')->value();
        $html = Storage::disk('s3')->get($request->string('content_key')->value());
        $page = new ArticlePage($url, $html);

        $article->update($page->toArray());

        return response()->noContent();
    }

    public function sitemap(ScrapeJobCompleteRequest $request, Bot $bot): Response
    {
        $url = $request->string('url')->value();
        $html = Storage::disk('s3')->get($request->string('content_key')->value());
        $page = Html::of($url, $html);

        (new UpdateOrCreateSitemapPage)(
            $bot,
            $url,
            $page->title(),
            $request->string('screenshot_key')->value(),
            $page->meta(),
        );

        /** @var array<string, bool> $known */
        $known = $bot->sitemap?->pages()->pluck('url')->mapWithKeys(fn (string $u) => [$u => true])->all() ?? [];

        (new UpdateSitemap)($bot, $known);

        // Continue the crawl: enqueue any not-yet-seen link found on this
        // page, one depth hop shallower. `$known` (backed by existing Page
        // rows) is this crawl's "done" set — the async equivalent of
        // SitemapBotJob's old in-memory $done array, just persisted so it
        // survives across separate worker invocations instead of one
        // function call's stack.
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
