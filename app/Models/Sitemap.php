<?php

namespace App\Models;

use App\Builders\SitemapBuilder;
use App\Observers\SitemapObserver;
use App\Traits\HasBot;
use App\Traits\HasPages;
use Database\Factories\SitemapFactory;
use Illuminate\Database\Eloquent\Attributes\Fillable;
use Illuminate\Database\Eloquent\Attributes\ObservedBy;
use Illuminate\Database\Eloquent\Attributes\UseEloquentBuilder;
use Illuminate\Database\Eloquent\Attributes\UseFactory;
use Illuminate\Database\Eloquent\Factories\HasFactory;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\SoftDeletes;

/**
 * @property int $id
 * @property string $domain
 * @property array<array-key, mixed>|null $urls
 * @property \Carbon\CarbonImmutable|null $created_at
 * @property \Carbon\CarbonImmutable|null $updated_at
 * @property \Carbon\CarbonImmutable|null $deleted_at
 * @property int $bot_id
 * @property-read \App\Models\Bot|null $bot
 * @property-read \Illuminate\Database\Eloquent\Collection<int, \App\Models\Page> $pages
 * @property-read int|null $pages_count
 * @method static SitemapBuilder<static>|Sitemap byDomain()
 * @method static \Database\Factories\SitemapFactory factory($count = null, $state = [])
 * @method static SitemapBuilder<static>|Sitemap newModelQuery()
 * @method static SitemapBuilder<static>|Sitemap newQuery()
 * @method static SitemapBuilder<static>|Sitemap notDeleted()
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Sitemap onlyTrashed()
 * @method static SitemapBuilder<static>|Sitemap query()
 * @method static SitemapBuilder<static>|Sitemap whereBotId($value)
 * @method static SitemapBuilder<static>|Sitemap whereCreatedAt($value)
 * @method static SitemapBuilder<static>|Sitemap whereDeletedAt($value)
 * @method static SitemapBuilder<static>|Sitemap whereDomain($value)
 * @method static SitemapBuilder<static>|Sitemap whereId($value)
 * @method static SitemapBuilder<static>|Sitemap whereUpdatedAt($value)
 * @method static SitemapBuilder<static>|Sitemap whereUrls($value)
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Sitemap withTrashed(bool $withTrashed = true)
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Sitemap withoutTrashed()
 * @mixin \Eloquent
 */
#[Fillable('bot_id', 'domain', 'urls')]
#[ObservedBy(SitemapObserver::class)]
#[UseEloquentBuilder(SitemapBuilder::class)]
#[UseFactory(SitemapFactory::class)]
class Sitemap extends Model
{
    /** @use HasFactory<SitemapFactory> */
    use HasBot, HasFactory, HasPages, SoftDeletes;

    /** @return array<string, string> */
    protected function casts(): array
    {
        return [
            'urls' => 'array',
        ];
    }
}
