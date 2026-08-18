<?php

namespace App\Models;

use App\Traits\HasBot;
use Closure;
use Database\Factories\ArticleFactory;
use Illuminate\Database\Eloquent\Attributes\Fillable;
use Illuminate\Database\Eloquent\Attributes\UseFactory;
use Illuminate\Database\Eloquent\Factories\HasFactory;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\SoftDeletes;
use Illuminate\Support\Arr;
use Illuminate\Support\Facades\URL;
use Illuminate\Support\Str;

/**
 * @property int $id
 * @property string $title
 * @property \Carbon\CarbonImmutable $published_at
 * @property string|null $img_alt
 * @property string|null $img_url
 * @property string|null $source
 * @property array<array-key, mixed>|null $keywords
 * @property string|null $description
 * @property string|null $body
 * @property \Carbon\CarbonImmutable|null $created_at
 * @property \Carbon\CarbonImmutable|null $updated_at
 * @property \Carbon\CarbonImmutable|null $deleted_at
 * @property int $bot_id
 * @property string|null $publisher
 * @property string $url
 * @property-read \App\Models\Bot|null $bot
 * @method static \Database\Factories\ArticleFactory factory($count = null, $state = [])
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Article newModelQuery()
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Article newQuery()
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Article onlyTrashed()
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Article query()
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Article whereBody($value)
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Article whereBotId($value)
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Article whereCreatedAt($value)
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Article whereDeletedAt($value)
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Article whereDescription($value)
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Article whereId($value)
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Article whereImgAlt($value)
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Article whereImgUrl($value)
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Article whereKeywords($value)
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Article wherePublishedAt($value)
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Article wherePublisher($value)
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Article whereSource($value)
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Article whereTitle($value)
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Article whereUpdatedAt($value)
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Article whereUrl($value)
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Article withTrashed(bool $withTrashed = true)
 * @method static \Illuminate\Database\Eloquent\Builder<static>|Article withoutTrashed()
 * @mixin \Eloquent
 */
#[Fillable([
    'body',
    'content',
    'description',
    'img_alt',
    'img_url',
    'keywords',
    'url',
    'published_at',
    'publisher',
    'source',
    'title',
])]
#[UseFactory(ArticleFactory::class)]
class Article extends Model
{
    /** @use HasFactory<ArticleFactory> */
    use HasBot, HasFactory, SoftDeletes;

    /** @return array<string, string> */
    protected function casts(): array
    {
        return [
            'keywords' => 'array',
            'published_at' => 'datetime',
        ];
    }

    public static function row(): Closure
    {
        return function (Article $article) {
            $source = str($article->title)
                ->afterLast(' - ')
                ->beforeLast('|')
                ->trim();

            $title = Str::beforeLast($article->title, ' - ');
            if ($source->is($title)) {
                $source = $article->publisher;
            } elseif ($source->substrCount(' ') > 2) {
                $source = str($source)->after('The')->trim();
                if ($source->substrCount(' ') > 2) {
                    $parts = $source->explode(' ');
                    $source = implode(' ', [$parts[0], $parts[1], $parts[2]]);
                }
            }

            if (blank($source)) {
                $source = URL::toDomain($article->url);
            }

            return [
                ...Arr::toCamelCase($article),
                ...[
                    'favicon' => URL::toFavicon($article->url),
                    'title' => $title,
                    'source' => $source,
                ],
            ];
        };
    }
}
