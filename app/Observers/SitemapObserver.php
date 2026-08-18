<?php

namespace App\Observers;

use App\Models\Sitemap;

class SitemapObserver
{
    public function deleting(Sitemap $model): void
    {
        $model->deletePages();
    }
}
