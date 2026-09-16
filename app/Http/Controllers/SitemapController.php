<?php

namespace App\Http\Controllers;

use App\Enums\BotType;
use App\Models\Bot;
use App\Models\Page;
use App\Models\Sitemap;
use Illuminate\Http\RedirectResponse;
use Illuminate\Http\Request;
use Illuminate\Routing\Attributes\Controllers\Authorize;
use Illuminate\Support\Facades\URL;
use Inertia\Inertia;
use Inertia\Response;

class SitemapController extends Controller
{
    public function index(Request $request): Response
    {
        return Inertia::render('sitemaps/Index', [
            'sitemaps' => Bot::query()
                ->user($request->user())
                ->type(BotType::Sitemap)
                ->get()
                ->map(fn (Bot $bot) => $bot->sitemap)
                ->sortBy('domain'),
        ]);
    }

    #[Authorize('delete', 'sitemap')]
    public function destroy(Sitemap $sitemap): RedirectResponse
    {
        $sitemap->delete();

        return to_route('sitemaps.index');
    }

    #[Authorize('view', 'sitemap')]
    public function show(Sitemap $sitemap): Response
    {
        return Inertia::render('sitemaps/Show', [
            'sitemap' => [
                'id' => $sitemap->id,
                'domain' => $sitemap->domain,
                'faviconUrl' => URL::toFavicon($sitemap->domain, 32),
                'urls' => $sitemap->urls ?? [],
                'pages' => $sitemap->pages()
                    ->get(['id', 'url', 'title', 'meta', 'screenshot_key'])
                    ->map(fn (Page $page) => [
                        'id' => $page->id,
                        'url' => $page->url,
                        'title' => $page->title,
                        'meta' => $page->meta ?? [],
                        'screenshotUrl' => $page->screenshot_key === null
                            ? null
                            : $page->screenshotUrl(),
                    ])
                    ->all(),
            ],
        ]);
    }
}
