<?php

namespace App\Traits;

use App\Models\Page;
use Illuminate\Database\Eloquent\Relations\MorphMany;

trait HasPages
{
    /** @return MorphMany<Page, $this> */
    public function pages(): MorphMany
    {
        return $this->MorphMany(Page::class, 'pageable');
    }

    public function deletePages(): void
    {
        $this->pages()->each(fn (Page $page) => $page->delete());
    }
}
