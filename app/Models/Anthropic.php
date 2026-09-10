<?php

namespace App\Models;

use App\Traits\HasUser;
use Carbon\CarbonImmutable;
use Illuminate\Database\Eloquent\Attributes\Fillable;
use Illuminate\Database\Eloquent\Model;

/**
 * @property int $id
 * @property int $user_id
 * @property string $api_key
 * @property string|null $default_model
 * @property CarbonImmutable|null $created_at
 * @property CarbonImmutable|null $updated_at
 * @property-read User|null $user
 *
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Anthropic newModelQuery()
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Anthropic newQuery()
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Anthropic query()
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Anthropic whereApiKey($value)
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Anthropic whereCreatedAt($value)
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Anthropic whereDefaultModel($value)
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Anthropic whereId($value)
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Anthropic whereUpdatedAt($value)
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Anthropic whereUserId($value)
 *
 * @mixin \Eloquent
 */
#[Fillable('api_key', 'default_model')]
class Anthropic extends Model
{
    use HasUser;
}
