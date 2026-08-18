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
