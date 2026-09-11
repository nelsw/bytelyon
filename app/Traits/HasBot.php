<?php

namespace App\Traits;

use App\Models\Bot;
use Illuminate\Database\Eloquent\Relations\BelongsTo;

trait HasBot
{
    /** @return BelongsTo<Bot, $this> */
    public function bot(): BelongsTo
    {
        return $this->belongsTo(Bot::class);
    }
}
