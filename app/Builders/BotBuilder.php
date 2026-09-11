<?php

namespace App\Builders;

use App\Enums\BotType;
use App\Models\Bot;
use Illuminate\Database\Eloquent\Builder;

/** @extends Builder<Bot> */
class BotBuilder extends Builder
{
    public function enabled(): static
    {
        return $this->where('enabled', true);
    }

    public function ready(): static
    {
        return $this->whereRaw("last_run_at IS NULL
OR (frequency = 'hourly' AND (last_run_at + interval '1 hour') < NOW())
OR (frequency = 'daily' AND (last_run_at + interval '1 day') < NOW())
OR (frequency = 'weekly' AND (last_run_at + interval '7 day') < NOW())
OR (frequency = 'monthly' AND (last_run_at + interval '30 day') < NOW())");
    }
}
