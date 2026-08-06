<?php

/** @noinspection LaravelUnknownRouteNameInspection */

namespace App\Http\Controllers;

use App\Concerns\BotValidationRules;
use App\Concerns\FiltersBots;
use App\Enums\BotType;
use App\Enums\FrequencyType;
use App\Models\Bot;
use Illuminate\Http\RedirectResponse;
use Illuminate\Http\Request;
use Illuminate\Routing\Attributes\Controllers\Authorize;
use Inertia\Inertia;
use Inertia\Response;

class BotController extends Controller
{
    use BotValidationRules;
    use FiltersBots;

    public function index(Request $request): Response
    {
        return Inertia::render('Dashboard', $this->paginatedBots($request));
    }

    #[Authorize('view', 'bot')]
    public function show(Bot $bot): Response
    {
        return Inertia::render('bots/Show', ['bot' => $bot]);
    }

    public function create(): Response
    {
        return Inertia::render('bots/Create', [
            'typeOptions' => BotType::options(),
            'frequencyOptions' => FrequencyType::options(),
        ]);
    }

    #[Authorize('update', 'bot')]
    public function edit(Bot $bot): Response
    {
        return Inertia::render('bots/Edit', [
            'bot' => $bot,
            'typeOptions' => BotType::options(),
            'frequencyOptions' => FrequencyType::options(),
        ]);
    }

    public function store(Request $request): RedirectResponse
    {
        $request->user()->bots()->create($request->validate($this->createRules(), [
            'query.unique' => __('You already have a bot with this query and type.'),
        ]));

        Inertia::flash('toast', ['type' => 'success', 'message' => __('Bot created.')]);

        return to_route('dashboard');
    }

    #[Authorize('update', 'bot')]
    public function update(Request $request, Bot $bot): RedirectResponse
    {
        $bot->update($request->validate($this->updateRules()));

        Inertia::flash('toast', ['type' => 'success', 'message' => __('Bot updated.')]);

        return to_route('dashboard');
    }

    #[Authorize('delete', 'bot')]
    public function destroy(Bot $bot): RedirectResponse
    {
        $bot->delete();

        Inertia::flash('toast', ['type' => 'success', 'message' => __('Bot deleted.')]);

        return to_route('dashboard');
    }
}
