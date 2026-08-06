<?php

namespace App\Traits;

use Illuminate\Support\Facades\Storage;

trait HasScreenshot
{
    public function screenshotUrl(): ?string
    {
        if ($this->screenshot_key === null) {
            return null;
        }

        $ttl = now()->addMinutes(15);

        return cache()->remember(
            $this->screenshot_key,
            $ttl,
            fn () => Storage::disk('s3')->temporaryUrl($this->screenshot_key, $ttl),
        );
    }

    public function deleteScreenshot(): void
    {
        if ($this->screenshot_key !== null) {
            Storage::disk('s3')->delete($this->screenshot_key);
        }
    }
}
