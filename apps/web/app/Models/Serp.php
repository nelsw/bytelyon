<?php

namespace App\Models;

use App\Builders\SerpBuilder;
use App\Observers\SerpObserver;
use App\Traits\HasBot;
use App\Traits\HasPages;
use App\Traits\HasScreenshot;
use Database\Factories\SerpFactory;
use Illuminate\Database\Eloquent\Attributes\Fillable;
use Illuminate\Database\Eloquent\Attributes\ObservedBy;
use Illuminate\Database\Eloquent\Attributes\UseEloquentBuilder;
use Illuminate\Database\Eloquent\Attributes\UseFactory;
use Illuminate\Database\Eloquent\Factories\HasFactory;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\SoftDeletes;

/**
 * @method static SerpBuilder query()
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
