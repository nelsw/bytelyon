<?php

namespace App\Models;

use App\Observers\PageObserver;
use App\Traits\HasScreenshot;
use Database\Factories\PageFactory;
use Illuminate\Database\Eloquent\Attributes\Fillable;
use Illuminate\Database\Eloquent\Attributes\ObservedBy;
use Illuminate\Database\Eloquent\Attributes\UseFactory;
use Illuminate\Database\Eloquent\Factories\HasFactory;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\MorphTo;
use Illuminate\Database\Eloquent\SoftDeletes;

/**
 * @property int $id
 * @property string $url
 * @property string $domain
 * @property string $title
 * @property string|null $screenshot_key
 * @property array<array-key, mixed>|null $meta
 * @property \Carbon\CarbonImmutable|null $created_at
 * @property \Carbon\CarbonImmutable|null $updated_at
 * @property \Carbon\CarbonImmutable|null $deleted_at
 * @property string $pageable_type
 * @property int $pageable_id
 * @property string|null $kind
 * @property int|null $index
 * @property-read Model|\Eloquent $pageable
 * @method static \Database\Factories\PageFactory factory($count = null, $state = [])
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Page newModelQuery()
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Page newQuery()
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Page onlyTrashed()
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Page query()
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Page whereCreatedAt($value)
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Page whereDeletedAt($value)
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Page whereDomain($value)
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Page whereId($value)
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Page whereIndex($value)
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Page whereKind($value)
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Page whereMeta($value)
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Page wherePageableId($value)
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Page wherePageableType($value)
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Page whereScreenshotKey($value)
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Page whereTitle($value)
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Page whereUpdatedAt($value)
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Page whereUrl($value)
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Page withTrashed(bool $withTrashed = true)
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Page withoutTrashed()
 * @mixin \Eloquent
 */
#[Fillable('domain', 'meta', 'screenshot_key', 'title', 'url', 'kind', 'index')]
#[ObservedBy(PageObserver::class)]
#[UseFactory(PageFactory::class)]
class Page extends Model
{
    /** @use HasFactory<PageFactory> */
    use HasFactory, HasScreenshot, SoftDeletes;

    protected $guarded = [
        'index',
        'kind',
    ];

    /** @return array<string, string> */
    protected function casts(): array
    {
        return [
            'meta' => 'array',
        ];
    }

    /** @noinspection PhpUnused */
    public function pageable(): MorphTo
    {
        return $this->morphTo();
    }
}
