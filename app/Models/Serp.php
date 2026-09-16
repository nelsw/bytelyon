<?php

namespace App\Models;

use App\Observers\SerpObserver;
use App\Traits\HasBot;
use App\Traits\HasPages;
use App\Traits\HasScreenshot;
use Carbon\CarbonImmutable;
use Database\Factories\SerpFactory;
use Eloquent;
use Illuminate\Database\Eloquent\Attributes\Fillable;
use Illuminate\Database\Eloquent\Attributes\ObservedBy;
use Illuminate\Database\Eloquent\Attributes\UseFactory;
use Illuminate\Database\Eloquent\Builder;
use Illuminate\Database\Eloquent\Collection;
use Illuminate\Database\Eloquent\Factories\HasFactory;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\SoftDeletes;

/**
 * @property int $id
 * @property string $query
 * @property string|null $screenshot_key
 * @property array<array-key, mixed>|null $data
 * @property CarbonImmutable|null $created_at
 * @property CarbonImmutable|null $updated_at
 * @property CarbonImmutable|null $deleted_at
 * @property int $bot_id
 * @property string|null $content_key
 * @property-read Collection<int, Page> $pages
 * @property-read int|null $pages_count
 *
 * @method static SerpFactory factory($count = null, $state = [])
 * @method static Builder<Serp|static> byQuery()
 * @method static Builder<Serp|static> newModelQuery()
 * @method static Builder<Serp|static> newQuery()
 * @method static Builder<Serp|static> notDeleted()
 * @method static Builder<Serp|static> onlyTrashed()
 * @method static Builder<Serp|static> query()
 * @method static Builder<Serp|static> whereBotId($value)
 * @method static Builder<Serp|static> whereQuery($value)
 * @method static Builder<Serp|static> whereScreenshotKey($value)
 * @method static Builder<Serp|static> whereUpdatedAt($value)
 * @method static Builder<Serp|static> withoutTrashed()
 * @method static Builder<Serp|static> withTrashed(bool $withTrashed = true)
 *
 * @mixin Eloquent
 */
#[Fillable('query', 'data', 'screenshot_key', 'content_key')]
#[ObservedBy(SerpObserver::class)]
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
