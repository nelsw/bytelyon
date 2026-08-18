<?php

namespace App\Models;

use App\Builders\BotBuilder;
use App\Enums\BotType;
use App\Enums\FrequencyType;
use App\Observers\BotObserver;
use App\Policies\BotPolicy;
use App\Traits\HasUser;
use Database\Factories\BotFactory;
use Illuminate\Database\Eloquent\Attributes\Fillable;
use Illuminate\Database\Eloquent\Attributes\ObservedBy;
use Illuminate\Database\Eloquent\Attributes\UseEloquentBuilder;
use Illuminate\Database\Eloquent\Attributes\UseFactory;
use Illuminate\Database\Eloquent\Attributes\UsePolicy;
use Illuminate\Database\Eloquent\Collection;
use Illuminate\Database\Eloquent\Factories\HasFactory;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\HasMany;
use Illuminate\Database\Eloquent\Relations\HasOne;
use Illuminate\Database\Eloquent\SoftDeletes;

/**
 * @property-read Collection<int, Article> $articles
 * @property-read Serp|null $serp
 * @property-read Sitemap|null $sitemap
 * @property-read User|null $user
 * @method static BotBuilder query()
 * @property int $id
 * @property string|null $blacklist
 * @property bool $enabled
 * @property bool $headless
 * @property FrequencyType $frequency
 * @property string $query
 * @property BotType $type
 * @property \Carbon\CarbonImmutable|null $last_run_at
 * @property \Carbon\CarbonImmutable|null $created_at
 * @property \Carbon\CarbonImmutable|null $updated_at
 * @property \Carbon\CarbonImmutable|null $deleted_at
 * @property int $user_id
 * @property string|null $last_run_result
 * @property-read int|null $articles_count
 * @method static BotBuilder<static>|Bot enabled(bool $b = true)
 * @method static \Database\Factories\BotFactory factory($count = null, $state = [])
 * @method static BotBuilder<static>|Bot headless(bool $b = true)
 * @method static BotBuilder<static>|Bot newModelQuery()
 * @method static BotBuilder<static>|Bot newQuery()
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Bot onlyTrashed()
 * @method static BotBuilder<static>|Bot ready()
 * @method static BotBuilder<static>|Bot type(\App\Enums\BotType|string $type)
 * @method static BotBuilder<static>|Bot whereBlacklist($value)
 * @method static BotBuilder<static>|Bot whereCreatedAt($value)
 * @method static BotBuilder<static>|Bot whereDeletedAt($value)
 * @method static BotBuilder<static>|Bot whereEnabled($value)
 * @method static BotBuilder<static>|Bot whereFrequency($value)
 * @method static BotBuilder<static>|Bot whereHeadless($value)
 * @method static BotBuilder<static>|Bot whereId($value)
 * @method static BotBuilder<static>|Bot whereLastRunAt($value)
 * @method static BotBuilder<static>|Bot whereLastRunResult($value)
 * @method static BotBuilder<static>|Bot whereQuery($value)
 * @method static BotBuilder<static>|Bot whereType($value)
 * @method static BotBuilder<static>|Bot whereUpdatedAt($value)
 * @method static BotBuilder<static>|Bot whereUserId($value)
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Bot withTrashed(bool $withTrashed = true)
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Bot withoutTrashed()
 * @mixin \Eloquent
 */
#[Fillable('enabled', 'frequency', 'query', 'type', 'last_run_at', 'headless')]
#[UseFactory(BotFactory::class)]
#[UsePolicy(BotPolicy::class)]
#[UseEloquentBuilder(BotBuilder::class)]
#[ObservedBy(BotObserver::class)]
class Bot extends Model
{
    /** @use HasFactory<BotFactory> */
    use HasFactory, HasUser, SoftDeletes;

    /** @return array<string, string> */
    protected function casts(): array
    {
        return [
            'frequency' => FrequencyType::class,
            'last_run_at' => 'datetime',
            'type' => BotType::class,
        ];
    }

    /** @return HasMany<Article, $this> */
    public function articles(): HasMany
    {
        return $this->hasMany(Article::class);
    }

    /** @return HasOne<Serp, $this> */
    public function serp(): HasOne
    {
        return $this->hasOne(Serp::class);
    }

    /** @return HasOne<Sitemap, $this> */
    public function sitemap(): HasOne
    {
        return $this->hasOne(Sitemap::class);
    }

    public function toArray(): array
    {
        return [
            'id' => $this->id,
            'type' => $this->type,
            'query' => $this->query,
            'enabled' => $this->enabled,
            'frequency' => $this->frequency,
            'blacklist' => $this->blacklist,
            'headless' => $this->headless,
            'processedAt' => $this->last_run_at,
            'createdAt' => $this->created_at,
            'updatedAt' => $this->updated_at,
            'childId' => match ($this->type) {
                BotType::Search => $this->serp?->id ?? 0,
                BotType::Sitemap => $this->sitemap?->id ?? 0,
                default => -1,
            },
            'pageCount' => match ($this->type) {
                BotType::News => $this->articles->count(),
                BotType::Search => $this->serp?->pages?->count(),
                BotType::Sitemap => $this->sitemap?->pages?->count(),
            },
        ];
    }

    public function toJson($options = 0): string
    {
        return json_encode([
            'id' => $this->id,
            'type' => $this->type,
            'query' => $this->query,
            'blacklist' => explode("\n", $this->blacklist),
            'headless' => $this->headless,
            'last_run_at' => ($this->last_run_at ?? now()->subYear()),
        ], $options);
    }
}
