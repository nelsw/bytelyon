<?php

namespace App\Traits;

use App\Facades\S3;

trait HasScreenshot
{
    public function screenshotUrl(): ?string
    {
        return S3::url($this->screenshot_key);
    }

    public function deleteScreenshot(): void
    {
        S3::del($this->screenshot_key);
    }
}
