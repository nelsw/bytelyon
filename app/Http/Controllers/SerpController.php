<?php

namespace App\Http\Controllers;

use App\Models\Serp;
use Illuminate\Http\RedirectResponse;
use Illuminate\Support\Facades\URL;
use Inertia\Inertia;
use Inertia\Response;

class SerpController extends Controller
{
    public function index(): Response
    {
        return Inertia::render('serps/Index', [
            'serps' => Serp::query()
                ->notDeleted()
                ->byQuery()
                ->withCount('pages')
                ->with('bot')
                ->get(),
        ]);
    }

    public function destroy(Serp $serp): RedirectResponse
    {
        $serp->delete();

        return to_route('serps.index');
    }

    public function show(Serp $serp): Response
    {
        return Inertia::render('serps/Show', [
            'serp' => [
                'id' => $serp->id,
                'query' => $serp->query,
                'data' => $serp->data ?? [],
                'similarQueries' => $serp->data['similar_queries'] ?? [],
                'screenshotUrl' => $serp->screenshotUrl(),
                'pages' => $serp->pages()
                    ->get(['id', 'kind', 'index', 'title', 'url', 'domain', 'created_at', 'screenshot_key', 'meta'])
                    ->map(fn ($page) => [
                        'id' => $page->id,
                        'faviconUrl' => URL::toFavicon($page->url, 32),
                        'kind' => str($page->kind)->replace('_', ' ')->title(),
                        'index' => $page->index,
                        'title' => $page->title,
                        'url' => $page->url,
                        'domain' => $page->domain,
                        'created_at' => $page->created_at,
                        'meta' => $page->meta ?? [],
                        'screenshotUrl' => $page->screenshotUrl(),
                    ])
                    ->all(),
            ],
        ]);
    }
}
