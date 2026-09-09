<?php

namespace App\Casts;

use App\ValueObjects\Image;
use Illuminate\Contracts\Database\Eloquent\CastsAttributes;
use Illuminate\Database\Eloquent\Model;

class AsImage implements CastsAttributes
{
    public function get(Model $model, string $key, mixed $value, array $attributes): Image
    {
        return new Image(
            $value['url'],
            $value['alt'],
        );
    }

    public function set(Model $model, string $key, mixed $value, array $attributes): array
    {
        return [
            'url' => $value->url,
            'alt' => $value->alt,
        ];
    }
}
