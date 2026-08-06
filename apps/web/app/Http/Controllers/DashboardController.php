<?php

namespace App\Http\Controllers;

use App\Concerns\FiltersBots;
use App\Models\Article;
use App\Models\Bot;
use App\Models\Serp;
use App\Models\Sitemap;
use Carbon\Constants\DiffOptions;
use Illuminate\Http\Request;
use Inertia\Inertia;
use Inertia\Response;

class DashboardController extends Controller
{
    use FiltersBots;

    public function index(Request $request): Response
    {
        $botIds = $request->user()->bots()->pluck('id');

        $recentBots = Bot::query()
            ->whereIn('id', $botIds)
            ->withCount('articles')
            ->orderByDesc('last_run_at')
            ->orderByDesc('created_at')
            ->limit(5)
            ->get();

        return Inertia::render('Dashboard', [
            ...$this->paginatedBots($request),
            'metrics' => [
                'bots' => $botIds->count(),
                'articles' => Article::query()->whereIn('bot_id', $botIds)->count(),
                'searches' => Serp::query()->whereIn('bot_id', $botIds)->count(),
                'sitemaps' => Sitemap::query()->whereIn('bot_id', $botIds)->count(),
            ],
            'recentBots' => $recentBots->map(fn (Bot $bot) => [
                'id' => $bot->id,
                'query' => $bot->query,
                'enabled' => $bot->enabled,
                'articles' => $bot->articles_count,
                'lastRunAt' => $bot->last_run_at?->since(other: now(), syntax: DiffOptions::DIFF_RELATIVE_TO_NOW) ?? 'Never',
            ])->all(),
        ]);
    }
}
