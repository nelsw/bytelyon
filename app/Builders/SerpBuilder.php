<?php

namespace App\Builders;

use App\Models\Serp;
use Illuminate\Database\Eloquent\Builder;

/** @extends Builder<Serp> */
class SerpBuilder extends Builder
{
    public function byQuery(): static
    {
        return $this->orderBy('query');
    }

    public function notDeleted(): static
    {
        return $this->whereNull('deleted_at');
    }
}
