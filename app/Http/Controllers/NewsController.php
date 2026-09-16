<?php

namespace App\Http\Controllers;

use App\Enums\BotType;
use App\Models\Bot;
use Illuminate\Http\Request;
use Inertia\Inertia;
use Inertia\Response;

class NewsController extends Controller
{
    public function index(Request $request): Response
    {
        return Inertia::render('news/Index', [
            'bots' => Bot::query()
                ->user($request->user())
                ->type(BotType::News)
                ->withCount('articles')
                ->orderBy('query')
                ->get(),
        ]);
    }
}
