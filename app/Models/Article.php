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
