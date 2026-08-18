<?php

namespace App\Http\Middleware;

use App\Enums\BotType;
use App\Enums\FrequencyType;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Gate;
use Inertia\Middleware;

class HandleInertiaRequests extends Middleware
{
    /**
     * @see https://inertiajs.com/server-side-setup#root-template
     *
     * @var string
     */
    protected $rootView = 'app';

    /**
     * @see https://inertiajs.com/asset-versioning
     */
    public function version(Request $request): ?string
    {
        return parent::version($request);
    }

    /**
     * @see https://inertiajs.com/shared-data
     *
     * @return array<string, mixed>
     */
    public function share(Request $request): array
    {
        return [
            ...parent::share($request),
            'name' => config('app.name'),
            'auth' => [
                'user' => $request->user(),
                'canViewHorizon' => Gate::allows('viewHorizon', [$request->user()]),
                'canViewTelescope' => Gate::allows('viewTelescope', [$request->user()]),
            ],
            'sidebarOpen' => ! $request->hasCookie('sidebar_state') || $request->cookie('sidebar_state') === 'true',
            'typeOptions' => BotType::options(),
            'frequencyOptions' => FrequencyType::options(),
        ];
    }
}
