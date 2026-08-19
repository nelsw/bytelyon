<?php

namespace App\Models;

use App\Traits\HasUser;
use Carbon\CarbonImmutable;
use Illuminate\Database\Eloquent\Attributes\Fillable;
use Illuminate\Database\Eloquent\Attributes\Table;
use Illuminate\Database\Eloquent\Model;

/**
 * @property int $id
 * @property int $user_id
 * @property string $client_id
 * @property string $client_secret
 * @property string|null $default_author_name
 * @property string|null $default_blog_id
 * @property string $store
 * @property CarbonImmutable|null $created_at
 * @property CarbonImmutable|null $updated_at
 * @property-read User|null $user
 *
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Shopify newModelQuery()
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Shopify newQuery()
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Shopify query()
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Shopify whereClientId($value)
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Shopify whereClientSecret($value)
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Shopify whereCreatedAt($value)
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Shopify whereDefaultAuthorName($value)
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Shopify whereDefaultBlogId($value)
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Shopify whereId($value)
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Shopify whereStore($value)
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Shopify whereUpdatedAt($value)
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Shopify whereUserId($value)
 *
 * @mixin \Eloquent
 */
#[Fillable('client_id', 'client_secret', 'default_author_name', 'default_blog_id', 'store')]
#[Table('shopifys')]
class Shopify extends Model
{
    use HasUser;
}
