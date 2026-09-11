<?php

namespace App\Providers;

use Illuminate\Support\Facades\Gate;
use Laravel\Horizon\Horizon;
use Laravel\Horizon\HorizonApplicationServiceProvider;

class HorizonServiceProvider extends HorizonApplicationServiceProvider
{
    public function boot(): void
    {
        parent::boot();

        Horizon::routeMailNotificationsTo('kowalski7012@gmail.com');
    }

    protected function authorization(): void
    {
        $this->gate();

        Horizon::auth(function ($request) {
            return Gate::check('viewHorizon', [$request->user()]);
        });
    }

    protected function gate(): void
    {
        Gate::define('viewHorizon', function ($user = null): bool {
            if ($this->app->environment('local')) {
                return true;
            }
            return optional($user)->email === 'kowalski7012@gmail.com';
        });
    }
}
