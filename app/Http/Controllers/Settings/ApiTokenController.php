<?php

namespace App\Http\Controllers\Settings;

use App\Http\Controllers\Controller;
use App\Http\Requests\Settings\ApiTokenStoreRequest;
use Illuminate\Http\RedirectResponse;
use Illuminate\Http\Request;
use Inertia\Inertia;
use Inertia\Response;
use Laravel\Sanctum\PersonalAccessToken;

class ApiTokenController extends Controller
{
    public function edit(Request $request): Response
    {
        return Inertia::render('settings/ApiTokens', [
            'tokens' => $request->user()
                ->tokens()
                ->latest()
                ->get()
                ->map(fn (PersonalAccessToken $token) => [
                    'id' => $token->id,
                    'name' => $token->name,
                    'created_at_diff' => $token->created_at->diffForHumans(),
                    'last_used_at_diff' => $token->last_used_at?->diffForHumans(),
                ])
                ->values()
                ->all(),
            'plainTextToken' => $request->session()->get('plainTextToken'),
        ]);
    }

    public function store(ApiTokenStoreRequest $request): RedirectResponse
    {
        $token = $request->user()->createToken($request->string('name')->value(), ['worker']);

        Inertia::flash('toast', ['type' => 'success', 'message' => __('API token issued.')]);

        return to_route('api-tokens.edit')->with('plainTextToken', $token->plainTextToken);
    }

    public function destroy(Request $request, int $token): RedirectResponse
    {
        /** @var PersonalAccessToken $token */
        $token = $request->user()->tokens()->whereKey($token)->firstOrFail();
        $token->delete();

        Inertia::flash('toast', ['type' => 'success', 'message' => __('API token revoked.')]);

        return to_route('api-tokens.edit');
    }
}
