<?php

namespace App\ValueObjects;

use App\Casts\AsImage;
use Illuminate\Contracts\Database\Eloquent\Castable;

class Image implements Castable
{
    public function __construct(
        public string $url,
        public string $alt,
    ){}

    public static function castUsing(array $arguments): string
    {
        return AsImage::class;
    }
}
