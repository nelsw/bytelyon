<?php

namespace App\Http\Middleware;

use Closure;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Log;
use Symfony\Component\HttpFoundation\Response as Status;

readonly class CheckApiKey
{
    private array $keys;

    public function __construct()
    {
        $this->keys = config('app.whitelist.keys');
    }

    public function handle(Request $request, Closure $next): JsonResponse
    {
        if (! $request->hasHeader('X-API-KEY')) {
            Log::warning('X-API-key is not provided');
            return response()->json(status: Status::HTTP_UNAUTHORIZED);
        }

        if (! in_array($request->header('X-API-KEY'), $this->keys)) {
            Log::warning("X-key {$request->header('X-API-KEY')} is not in whitelist");
            return response()->json(status: Status::HTTP_UNAUTHORIZED);
        }

        return $next($request);
    }
}
