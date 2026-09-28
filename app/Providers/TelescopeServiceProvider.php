<?php

namespace App\Providers;

use App\Models\User;
use Illuminate\Support\Facades\Gate;
use Laravel\Telescope\IncomingEntry;
use Laravel\Telescope\Telescope;
use Laravel\Telescope\TelescopeApplicationServiceProvider;

class TelescopeServiceProvider extends TelescopeApplicationServiceProvider
{
    public function register(): void
    {
        $this->hideSensitiveRequestDetails();

        // Always render Telescope's own dashboard with its dark Bootstrap
        // dist (see public/css/telescope-theme.css, which now assumes this
        // dist is always active and no longer has a light variant).
        Telescope::night();

        $isLocal = $this->app->environment('local');

        Telescope::filter(function (IncomingEntry $entry) use ($isLocal) {
            return $isLocal
                || $entry->isEvent()
                || $entry->hasMonitoredTag()
                || $entry->isFailedRequest()
                || $entry->isLog()
                || $entry->isReportableException()
                || $entry->isScheduledTask();
        });
    }

    protected function hideSensitiveRequestDetails(): void
    {
        if ($this->app->environment('local')) {
            return;
        }

        Telescope::hideRequestParameters(['_token']);

        Telescope::hideRequestHeaders([
            'cookie',
            'x-csrf-token',
            'x-xsrf-token',
        ]);
    }

    protected function gate(): void
    {
        Gate::define('viewTelescope', function (User $user) {
            return $this->app->environment('local') || $user->email === config('app.admin');
        });
    }
}
