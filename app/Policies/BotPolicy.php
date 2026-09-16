<?php

namespace App\Policies;

use App\Models\Bot;
use App\Models\User;

class BotPolicy
{
    public function create(User $user): bool
    {
        return true;
    }

    public function view(User $user, Bot $bot): bool
    {
        return $bot->user()->is($user) || $user->email === config('app.admin');
    }

    public function update(User $user, Bot $bot): bool
    {
        return $bot->user()->is($user);
    }

    public function delete(User $user, Bot $bot): bool
    {
        return $bot->user()->is($user) || $user->email === config('app.admin');
    }
}
