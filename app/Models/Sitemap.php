<?php

namespace App\Models;

use App\Observers\SitemapObserver;
use App\Traits\HasBot;
use App\Traits\HasPages;
use Carbon\CarbonImmutable;
use Database\Factories\SitemapFactory;
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
 * @property string $domain
 * @property array<array-key, mixed>|null $urls
 * @property CarbonImmutable|null $created_at
 * @property CarbonImmutable|null $updated_at
 * @property CarbonImmutable|null $deleted_at
 * @property int $bot_id
 * @property-read Bot|null $bot
 * @property-read Collection<int, Page> $pages
 * @property-read int|null $pages_count
 *
 * @method static Builder<Sitemap|static> query()
 * @method static SitemapFactory factory($count = null, $state = [])
 *
 * @mixin Eloquent
 */
#[Fillable('domain', 'urls')]
#[ObservedBy(SitemapObserver::class)]
#[UseFactory(SitemapFactory::class)]
class Sitemap extends Model
{
    /** @use HasFactory<SitemapFactory> */
    use HasBot,
        HasFactory,
        HasPages,
        SoftDeletes;

    /** @return array<string, string> */
    protected function casts(): array
    {
        return [
            'urls' => 'array',
        ];
    }
}
