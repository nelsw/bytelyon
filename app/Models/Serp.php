<?php

namespace App\Models;

use App\Builders\SerpBuilder;
use App\Observers\SerpObserver;
use App\Traits\HasBot;
use App\Traits\HasPages;
use App\Traits\HasScreenshot;
use Carbon\CarbonImmutable;
use Database\Factories\SerpFactory;
use Illuminate\Database\Eloquent\Attributes\Fillable;
use Illuminate\Database\Eloquent\Attributes\ObservedBy;
use Illuminate\Database\Eloquent\Attributes\UseEloquentBuilder;
use Illuminate\Database\Eloquent\Attributes\UseFactory;
use Illuminate\Database\Eloquent\Collection;
use Illuminate\Database\Eloquent\Factories\HasFactory;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\SoftDeletes;

/**
 * @method static SerpBuilder query()
 *
 * @property int $id
 * @property string $query
 * @property string|null $screenshot_key
 * @property array<array-key, mixed>|null $data
 * @property CarbonImmutable|null $created_at
 * @property CarbonImmutable|null $updated_at
 * @property CarbonImmutable|null $deleted_at
 * @property int $bot_id
 * @property string|null $content_key
 * @property-read Bot|null $bot
 * @property-read Collection<int, Page> $pages
 * @property-read int|null $pages_count
 *
 * @method static SerpBuilder<static>|Serp byQuery()
 * @method static \Database\Factories\SerpFactory factory($count = null, $state = [])
 * @method static SerpBuilder<static>|Serp newModelQuery()
 * @method static SerpBuilder<static>|Serp newQuery()
 * @method static SerpBuilder<static>|Serp notDeleted()
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Serp onlyTrashed()
 * @method static SerpBuilder<static>|Serp whereBotId($value)
 * @method static SerpBuilder<static>|Serp whereContentKey($value)
 * @method static SerpBuilder<static>|Serp whereCreatedAt($value)
 * @method static SerpBuilder<static>|Serp whereData($value)
 * @method static SerpBuilder<static>|Serp whereDeletedAt($value)
 * @method static SerpBuilder<static>|Serp whereId($value)
 * @method static SerpBuilder<static>|Serp whereQuery($value)
 * @method static SerpBuilder<static>|Serp whereScreenshotKey($value)
 * @method static SerpBuilder<static>|Serp whereUpdatedAt($value)
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Serp withTrashed(bool $withTrashed = true)
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Serp withoutTrashed()
 *
 * @mixin \Eloquent
 */
#[Fillable('query', 'data', 'screenshot_key', 'content_key')]
#[ObservedBy(SerpObserver::class)]
#[UseEloquentBuilder(SerpBuilder::class)]
#[UseFactory(SerpFactory::class)]
class Serp extends Model
{
    /** @use HasFactory<SerpFactory> */
    use HasBot,
        HasFactory,
        HasPages,
        HasScreenshot,
        SoftDeletes;

    /** @return array<string, string> */
    protected function casts(): array
    {
        return [
            'data' => 'json',
        ];
    }
}
