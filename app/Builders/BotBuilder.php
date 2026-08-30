<?php

namespace App\Builders;

use App\Enums\BotType;
use App\Models\Bot;
use Illuminate\Database\Eloquent\Builder;

/** @extends Builder<Bot> */
class BotBuilder extends Builder
{
    public function type(BotType|string $type): static
    {
        return $this->where('type', $type);
    }

    public function enabled(bool $b = true): static
    {
        return $this->where('enabled', $b);
    }

    public function headless(bool $b = true): static
    {
        return $this->where('headless', $b);
    }

    public function user(int $id): static
    {
        return $this->where('user_id', $id);
    }

    public function ready(): static
    {
        return $this->whereRaw("played_at IS NULL
OR (frequency = 'hourly' AND (played_at + interval '1 hour') < NOW())
OR (frequency = 'daily' AND (played_at + interval '1 day') < NOW())
OR (frequency = 'weekly' AND (played_at + interval '7 day') < NOW())
OR (frequency = 'monthly' AND (played_at + interval '30 day') < NOW())");
    }
}
