<?php

namespace App\Http\Controllers;

use App\Concerns\ArticleValidationRules;
use App\Concerns\PageValidationRules;
use App\Concerns\SerpValidationRules;
use App\Concerns\SitemapValidationRules;
use App\Models\Bot;
use App\Models\Page;
use App\Models\Serp;
use App\Models\Sitemap;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Support\Arr;
use Illuminate\Support\Facades\Log;
use Illuminate\Support\Facades\Redis;
use Illuminate\Support\Facades\Storage;
use Illuminate\Support\Facades\URL;

class ApiController extends Controller
{
    use ArticleValidationRules,
        PageValidationRules,
        SerpValidationRules,
        SitemapValidationRules;

    public function bots(): JsonResponse
    {
        return response()->json(data: Bot::query()
            ->user(auth()->id())
            ->enabled()
            ->ready()
            ->get());
    }

    public function bot(Bot $bot): JsonResponse
    {
        $bot->save();
        return response()->json();
    }

    public function article(Request $request, Bot $bot): JsonResponse
    {
        $bot->articles()->updateOrCreate(
            attributes: ['url' => $request->input('url')],
            values: $request->validate($this->articleRules($bot->last_run_at)),
        );
        return response()->json();
    }

    public function serp(Request $request, Bot $bot): JsonResponse
    {
        return response()->json([
            'id' => $bot->serp()->updateOrCreate(
                attributes: ['query' => $request->input('query')],
                values: $request->validate($this->serpRules()),
            )->getKey(),
        ]);
    }

    public function sitemap(Request $request, Bot $bot): JsonResponse
    {
        return response()->json([
            'id' => $bot->sitemap()->updateOrCreate(
                attributes: ['domain' => $request->input('domain')],
                values: $request->validate($this->sitemapRules()),
            )->getKey(),
        ]);
    }

    public function serpPage(Request $request, Serp $serp): JsonResponse
    {
        return response()->json([
            'id' => $serp->pages()->updateOrCreate(
                attributes: ['url' => $request->input('url')],
                values: $request->validate($this->pageRules()),
            )->getKey(),
        ]);
    }

    public function sitemapPage(Request $request, Sitemap $sitemap): JsonResponse
    {
        return response()->json([
            'id' => $sitemap->pages()->updateOrCreate(
                attributes: ['url' => $request->input('url')],
                values: $request->validate($this->pageRules()),
            )->getKey(),
        ]);
    }

    public function pageImg(Request $request, Page $page): JsonResponse
    {
        Storage::disk('s3')->putFile(
            path: $page->screenshot_key,
            file: $request->validate(['screenshot_data' => ['nullable', 'image', 'mimes:png']])['screenshot_data'],
        );
        return response()->json();
    }

    public function searchImg(Request $request, Serp $serp): JsonResponse
    {
        Storage::disk('s3')->putFile(
            path: $serp->screenshot_key,
            file: $request->validate(['screenshot_data' => ['nullable', 'image', 'mimes:png']])['screenshot_data'],
        );
        return response()->json();
    }
}
