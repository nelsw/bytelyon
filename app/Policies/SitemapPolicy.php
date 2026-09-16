<?php

namespace App\Policies;

use App\Models\Sitemap;
use App\Models\User;

class SitemapPolicy
{
    public function viewAny(User $user): bool
    {
        return $user->email === config('app.admin');
    }

    public function view(User $user, Sitemap $sitemap): bool
    {
        return $sitemap->bot->user()->is($user) || $user->email === config('app.admin');
    }

    public function create(User $user): bool
    {
        return true;
    }

    public function update(User $user, Sitemap $sitemap): bool
    {
        return $sitemap->bot->user()->is($user) || $user->email === config('app.admin');
    }

    public function delete(User $user, Sitemap $sitemap): bool
    {
        return $sitemap->bot->user()->is($user) || $user->email === config('app.admin');
    }

    public function restore(User $user, Sitemap $sitemap): bool
    {
        return false;
    }

    public function forceDelete(User $user, Sitemap $sitemap): bool
    {
        return false;
    }
}
