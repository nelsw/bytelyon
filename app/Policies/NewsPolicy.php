<?php

namespace App\Policies;

use App\Models\Article;
use App\Models\User;

class NewsPolicy
{
    public function create(User $user): bool
    {
        return true;
    }

    public function viewAny(User $user): bool
    {
        return true;
    }

    public function view(User $user, Article $article): bool
    {
        return $article->bot->user()->is($user) || $user->email === config('app.admin');
    }

    public function update(User $user, Article $article): bool
    {
        return $article->bot->user()->is($user) || $user->email === config('app.admin');
    }

    public function delete(User $user, Article $article): bool
    {
        return $article->bot->user()->is($user) || $user->email === config('app.admin');
    }
}
