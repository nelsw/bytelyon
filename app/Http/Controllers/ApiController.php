<?php

namespace App\Http\Controllers;

use App\Concerns\ArticleValidationRules;
use App\Concerns\PageValidationRules;
use App\Concerns\SerpValidationRules;
use App\Concerns\SitemapValidationRules;
use App\Models\Bot;
use App\Models\Serp;
use App\Models\Sitemap;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Support\Arr;
use Illuminate\Support\Facades\Redis;
use Illuminate\Support\Facades\URL;

class ApiController extends Controller
{
    use ArticleValidationRules,
        PageValidationRules,
        SerpValidationRules,
        SitemapValidationRules;

    public function bots(): JsonResponse
    {
        $bro = Redis::connection('broker');

        $keys = $bro->keys('bot:*:todo');
        if (! is_array($keys)) {
            return response()->json();
        }
        return response()->json(Arr::map($keys, fn (string $key) => json_decode($bro->getDel($key), true)));
    }

    public function bot(Request $request, Bot $bot): JsonResponse
    {
        $bro = Redis::connection('broker');
        $result=$request->input('result');
        $bro->set("bot:$bot->id:done", $result);
        if ($result === 'ok') {
            $bot->update(['last_run_at' => now()]);
        }
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
        $model = $bot->serp()->updateOrCreate(
            attributes: ['query' => $request->input('query')],
            values: $request->validate($this->serpRules()),
        );
        return response()->json(['id' => $model->id]);
    }

    public function sitemap(Request $request, Bot $bot): JsonResponse
    {
        $model = $bot->sitemap()->updateOrCreate(
            attributes: ['domain' => $request->input('domain')],
            values: $request->validate($this->sitemapRules()),
        );
        return response()->json(['id' => $model->id]);
    }

    public function serpPage(Request $request, Serp $serp): JsonResponse
    {
        $values = $request->validate($this->pageRules());
        if (! isset($values['domain'])) {
            $values['domain'] = URL::toDomain($values['url']);
        }
        $serp->pages()->updateOrCreate(['url' => $values['url']], $values);
        return response()->json();
    }

    public function sitemapPage(Request $request, Sitemap $sitemap): JsonResponse
    {
        $values = $request->validate($this->pageRules());
        if (! isset($values['domain'])) {
            $values['domain'] = URL::toDomain($values['url']);
        }
        $sitemap->pages()->updateOrCreate(['url' => $values['url']], $values);
        return response()->json();
    }
}
