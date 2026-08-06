<?php

namespace App\Builders;

use App\Models\Sitemap;
use Illuminate\Database\Eloquent\Builder;

/** @extends Builder<Sitemap> */
class SitemapBuilder extends Builder
{
    public function byDomain(): static
    {
        return $this->orderBy('domain');
    }

    public function notDeleted(): static
    {
        return $this->whereNull('deleted_at');
    }
}
