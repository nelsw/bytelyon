<?php

namespace App\Http\Controllers;

use App\Models\Serp;
use Illuminate\Http\RedirectResponse;
use Illuminate\Support\Arr;
use Illuminate\Support\Facades\URL;
use Inertia\Inertia;
use Inertia\Response;

class SerpController extends Controller
{
    public function index(): Response
    {
        return Inertia::render('serps/Index', [
            'serps' => Serp::query()
                ->orderBy('query')
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
        $data = collect($serp->data ?? []);

        return Inertia::render('serps/Show', [
            'serp' => [
                'id' => $serp->id,
                'query' => $serp->query,
                'similarQueries' => collect($data->get('similar_queries', []))
                    ->pluck('value')
                    ->filter()
                    ->values()
                    ->all(),
                'screenshotUrl' => $serp->screenshotUrl(),
                'results' => $data->except('similar_queries')
                    ->flatMap(fn (array $items, string $key) => collect($items)->map(fn (array $item) => [
                        'id' => "$key-".($item['index'] ?? 0),
                        'faviconUrl' => filled($item['url'] ?? null) ? URL::toFavicon($item['url'], 32) : null,
                        'kind' => str($item['kind'] ?? $key)->replace('_', ' ')->title()->toString(),
                        'index' => $item['index'] ?? null,
                        'title' => $item['title'] ?? '',
                        'url' => $item['url'] ?? null,
                        'domain' => $item['domain'] ?? null,
                        'meta' => Arr::except($item, ['kind', 'index', 'title', 'url', 'domain']),
                    ]))
                    ->values()
                    ->all(),
            ],
        ]);
    }
}
