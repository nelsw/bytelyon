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
