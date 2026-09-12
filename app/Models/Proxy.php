<?php

namespace App\Models;

use Carbon\CarbonImmutable;
use Database\Factories\ProxyFactory;
use Illuminate\Database\Eloquent\Attributes\Fillable;
use Illuminate\Database\Eloquent\Attributes\Table;
use Illuminate\Database\Eloquent\Attributes\UseFactory;
use Illuminate\Database\Eloquent\Builder;
use Illuminate\Database\Eloquent\Factories\HasFactory;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\BelongsTo;

/**
 * @property int $id
 * @property string $scheme
 * @property string $host
 * @property int|null $port
 * @property string|null $user
 * @property string|null $pass
 * @property string|null $bypass
 * @property int $user_id
 * @property string $name
 * @property CarbonImmutable|null $created_at
 * @property CarbonImmutable|null $updated_at
 * @property-read User|null $owner
 *
 * @method static ProxyFactory factory($count = null, $state = [])
 * @method static Builder<static>|Proxy newModelQuery()
 * @method static Builder<static>|Proxy newQuery()
 * @method static Builder<static>|Proxy query()
 * @method static Builder<static>|Proxy whereBypass($value)
 * @method static Builder<static>|Proxy whereCreatedAt($value)
 * @method static Builder<static>|Proxy whereHost($value)
 * @method static Builder<static>|Proxy whereId($value)
 * @method static Builder<static>|Proxy whereName($value)
 * @method static Builder<static>|Proxy wherePass($value)
 * @method static Builder<static>|Proxy wherePort($value)
 * @method static Builder<static>|Proxy whereScheme($value)
 * @method static Builder<static>|Proxy whereUpdatedAt($value)
 * @method static Builder<static>|Proxy whereUser($value)
 * @method static Builder<static>|Proxy whereUserId($value)
 */
#[Fillable('name', 'scheme', 'host', 'port', 'user', 'pass', 'bypass')]
#[UseFactory(ProxyFactory::class)]
#[Table('proxies')]
class Proxy extends Model
{
    /** @use HasFactory<ProxyFactory> */
    use HasFactory;

    /**
     * The proxy's owning user.
     *
     * Named `owner` (rather than the shared `HasUser::user()` relation)
     * because the `user` column holds the proxy's own auth username.
     *
     * @return BelongsTo<User, $this>
     */
    public function owner(): BelongsTo
    {
        return $this->belongsTo(User::class);
    }

    public function __toString(): string {
        return "$this->scheme://$this->user:$this->pass@$this->host:$this->port";
    }
    // todo - playwright specific array
}
