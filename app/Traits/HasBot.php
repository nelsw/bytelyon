<?php

namespace App\Traits;

use App\Enums\FrequencyType;
use App\Models\Bot;
use Carbon\CarbonInterface;
use Carbon\CarbonInterval;
use Illuminate\Database\Eloquent\Relations\BelongsTo;

/**
 * @property-read Bot|null $bot
 */
trait HasBot
{
    /** @return BelongsTo<Bot, $this> */
    public function bot(): BelongsTo
    {
        return $this->belongsTo(Bot::class);
    }

    public function isRunnable(): bool
    {
        return $this->bot->enabled && $this->bot->lastRunAt()->add(match ($this->bot->frequency) {
            FrequencyType::Hourly => CarbonInterval::hour(),
            FrequencyType::Daily => CarbonInterval::day(),
            FrequencyType::Weekly => CarbonInterval::week(),
            FrequencyType::Monthly => CarbonInterval::month(),
        })->isPast();
    }

    public function lastRunAt(): CarbonInterface
    {
        return $this->bot->last_run_at ?? now()->subYear();
    }
}
