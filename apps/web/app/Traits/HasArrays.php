<?php

namespace App\Traits;

use Illuminate\Support\Arr;

trait HasArrays
{
    public static function values(): array
    {
        return array_column(self::cases(), 'value');
    }

    public static function names(): array
    {
        return array_column(self::cases(), 'name');
    }

    public static function array(): array
    {
        return array_combine(self::values(), self::names());
    }

    public static function options(): array
    {
        return Arr::map(self::cases(), fn (self $type): array => [
            'value' => $type->value,
            'label' => $type->name,
        ]);
    }
}
