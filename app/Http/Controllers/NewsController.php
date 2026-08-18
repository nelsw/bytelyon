<?php

namespace App\Http\Controllers;

use App\Enums\BotType;
use Illuminate\Http\Request;
use Inertia\Inertia;
use Inertia\Response;

class NewsController extends Controller
{
    public function index(Request $request): Response
    {
        return Inertia::render('news/Index', [
            'bots' => $request->user()
                ->bots()
                ->type(BotType::News)
                ->withCount('articles')
                ->orderBy('query')
                ->get(),
        ]);
    }
}
