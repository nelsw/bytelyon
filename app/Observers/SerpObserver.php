<?php

namespace App\Observers;

use App\Facades\S3;
use App\Models\Serp;

class SerpObserver
{
    public function deleting(Serp $model): void
    {
        if ($model->content_key) {
            S3::del($model->content_key);
        }
        $model->deleteScreenshot();
        $model->deletePages();
    }
}
