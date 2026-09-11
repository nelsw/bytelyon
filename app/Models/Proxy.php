<?php

namespace App\Models;

use App\Traits\HasUser;
use Database\Factories\ProxyFactory;
use Illuminate\Database\Eloquent\Attributes\Fillable;
use Illuminate\Database\Eloquent\Attributes\Table;
use Illuminate\Database\Eloquent\Attributes\UseFactory;
use Illuminate\Database\Eloquent\Factories\HasFactory;
use Illuminate\Database\Eloquent\Model;

#[Fillable('name', 'server', 'username', 'password', 'bypass')]
#[UseFactory(ProxyFactory::class)]
#[Table('proxies')]
class Proxy extends Model
{
    /** @use HasFactory<ProxyFactory> */
    use HasFactory, HasUser;

    public function toPayload(): array
    {
        return [
            'server' => $this->server,
            'username' => $this->username,
            'password' => $this->password,
            'bypass' => $this->bypass,
        ];
    }
}
