<?php

namespace App\Traits;

use App\Enums\BotType;
use App\Models\Bot;
use Illuminate\Database\Eloquent\Relations\BelongsTo;

trait HasBot
{
    /** @return BelongsTo<Bot, $this> */
    public function bot(): BelongsTo
    {
        return $this->belongsTo(Bot::class);
    }

    public function type(): BotType
    {
        return $this->bot->type;
    }

    public function headless(): bool
    {
        return $this->bot->headless;
    }

    public function botType(): BotType
    {
        return $this->type;
    }
}
