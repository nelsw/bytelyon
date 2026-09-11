<?php

namespace App\Models;

use App\Traits\HasUser;
use Carbon\CarbonImmutable;
use Database\Factories\ProxyFactory;
use Illuminate\Database\Eloquent\Attributes\Fillable;
use Illuminate\Database\Eloquent\Attributes\Table;
use Illuminate\Database\Eloquent\Attributes\UseFactory;
use Illuminate\Database\Eloquent\Builder;
use Illuminate\Database\Eloquent\Factories\HasFactory;
use Illuminate\Database\Eloquent\Model;

/**
 * @property int $id
 * @property string $server
 * @property string $username
 * @property string $password
 * @property string|null $bypass
 * @property int $user_id
 * @property string $name
 * @property CarbonImmutable|null $created_at
 * @property CarbonImmutable|null $updated_at
 *
 * @property-read User|null $user
 *
 * @method static ProxyFactory factory($count = null, $state = [])
 * @method static Builder<static>|Proxy newModelQuery()
 * @method static Builder<static>|Proxy newQuery()
 * @method static Builder<static>|Proxy query()
 * @method static Builder<static>|Proxy whereBypass($value)
 * @method static Builder<static>|Proxy whereCreatedAt($value)
 * @method static Builder<static>|Proxy whereId($value)
 * @method static Builder<static>|Proxy whereName($value)
 * @method static Builder<static>|Proxy wherePassword($value)
 * @method static Builder<static>|Proxy whereServer($value)
 * @method static Builder<static>|Proxy whereUpdatedAt($value)
 * @method static Builder<static>|Proxy whereUserId($value)
 * @method static Builder<static>|Proxy whereUsername($value)
 */
#[Fillable('name', 'server', 'username', 'password', 'bypass')]
#[UseFactory(ProxyFactory::class)]
#[Table('proxies')]
class Proxy extends Model
{
    /** @use HasFactory<ProxyFactory> */
    use HasFactory, HasUser;
}
