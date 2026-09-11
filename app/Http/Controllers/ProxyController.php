<?php

namespace App\Http\Controllers;

use App\Http\Requests\ProxyRequest;
use App\Http\Resources\ProxyResource;
use App\Models\Proxy;
use Illuminate\Foundation\Auth\Access\AuthorizesRequests;

class ProxyController extends Controller
{
    use AuthorizesRequests;

    public function index()
    {
        $this->authorize('viewAny', Proxy::class);

        return ProxyResource::collection(Proxy::all());
    }

    public function store(ProxyRequest $request)
    {
        $this->authorize('create', Proxy::class);

        return new ProxyResource(Proxy::create($request->validated()));
    }

    public function show(Proxy $proxy)
    {
        $this->authorize('view', $proxy);

        return new ProxyResource($proxy);
    }

    public function update(ProxyRequest $request, Proxy $proxy)
    {
        $this->authorize('update', $proxy);

        $proxy->update($request->validated());

        return new ProxyResource($proxy);
    }

    public function destroy(Proxy $proxy)
    {
        $this->authorize('delete', $proxy);

        $proxy->delete();

        return response()->json();
    }
}
