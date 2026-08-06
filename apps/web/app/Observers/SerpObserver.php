<?php

namespace App\Observers;

use App\Models\Serp;
use Illuminate\Support\Facades\Storage;

class SerpObserver
{
    public function deleting(Serp $model): void
    {
        if ($model->content_key) {
            Storage::disk('s3')->delete($model->content_key);
        }
        $model->deleteScreenshot();
        $model->deletePages();
    }
}
