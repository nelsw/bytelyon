<?php

namespace App\Observers;

use App\Models\Page;

class PageObserver
{
    public function deleting(Page $model): void
    {
        $model->deleteScreenshot();
    }
}
