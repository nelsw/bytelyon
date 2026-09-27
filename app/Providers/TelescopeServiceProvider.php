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
        $this->configureDarkMode();

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

    /**
     * Match Telescope's own dashboard theme to the app-wide `appearance`
     * preference (see HandleAppearance middleware and
     * resources/js/composables/useAppearance.ts). Telescope::css() inlines
     * one of two fully-compiled Bootstrap stylesheets (dist/styles.css or
     * dist/styles-dark.css) chosen by the static Telescope::$useDarkTheme
     * flag, so — unlike the main app's `.dark` class toggle, which lets one
     * stylesheet react to both states — the choice has to be made here,
     * request-time, before that view renders, and can't be flipped
     * client-side afterwards without a full reload.
     *
     * The `appearance` cookie is read directly off the request rather than
     * via the shared view value, since provider registration runs before
     * the `web` middleware group (incl. HandleAppearance) executes; reading
     * it here is safe because bootstrap/app.php excludes `appearance` from
     * cookie encryption. A `system` preference (or no cookie yet) falls
     * back to the light theme, since the server has no way to know the
     * browser's OS-level preference on this request.
     */
    protected function configureDarkMode(): void
    {
        if (request()->cookie('appearance') === 'dark') {
            Telescope::night();
        }
    }

    protected function gate(): void
    {
        Gate::define('viewTelescope', function (User $user) {
            return $this->app->environment('local') || $user->email === config('app.admin');
        });
    }
}
