<?php

namespace App\Traits;

use Illuminate\Support\Uri;

trait HasUrl
{
    public function domain(): string
    {
        return Uri::domain($this->url);
    }
}
