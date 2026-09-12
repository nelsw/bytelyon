<?php

namespace App\Policies;

use App\Models\Proxy;
use App\Models\User;
use Illuminate\Auth\Access\HandlesAuthorization;

class ProxyPolicy
{
    use HandlesAuthorization;

    public function view(User $user, Proxy $proxy): bool
    {
        return $proxy->user_id === $user->id;
    }

    public function create(User $user): bool
    {
        return true;
    }

    public function update(User $user, Proxy $proxy): bool
    {
        return $proxy->user_id === $user->id;
    }

    public function delete(User $user, Proxy $proxy): bool
    {
        return $proxy->user_id === $user->id;
    }
}
