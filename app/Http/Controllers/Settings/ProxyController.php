<?php

namespace App\Http\Controllers\Settings;

use App\Http\Controllers\Controller;
use App\Http\Requests\Settings\ProxyStoreRequest;
use App\Models\Proxy;
use Illuminate\Foundation\Auth\Access\AuthorizesRequests;
use Illuminate\Http\RedirectResponse;
use Illuminate\Http\Request;
use Inertia\Inertia;
use Inertia\Response;

class ProxyController extends Controller
{
    use AuthorizesRequests;

    public function edit(Request $request): Response
    {
        return Inertia::render('settings/Proxies', [
            'proxies' => $request->user()
                ->proxies()
                ->latest()
                ->get()
                ->map(fn (Proxy $proxy) => [
                    'id' => $proxy->id,
                    'name' => $proxy->name,
                    'protocol' => $proxy->protocol,
                    'server' => $proxy->server,
                    'port' => $proxy->port,
                    'username' => $proxy->username,
                    'bypass' => $proxy->bypass,
                    'created_at_diff' => $proxy->created_at?->diffForHumans(),
                ])
                ->values()
                ->all(),
        ]);
    }

    public function store(ProxyStoreRequest $request): RedirectResponse
    {
        $this->authorize('create', Proxy::class);

        $request->user()->proxies()->create($request->validated());

        Inertia::flash('toast', ['type' => 'success', 'message' => __('Proxy added.')]);

        return to_route('proxies.edit');
    }

    public function destroy(Proxy $proxy): RedirectResponse
    {
        $this->authorize('delete', $proxy);

        $proxy->delete();

        Inertia::flash('toast', ['type' => 'success', 'message' => __('Proxy deleted.')]);

        return to_route('proxies.edit');
    }
}
