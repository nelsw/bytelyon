<?php

namespace App\Contracts;

use Illuminate\Contracts\Support\Arrayable;

interface Pageable extends Arrayable
{
    public function body(): string;

    public function description(): string;

    public function domain(): string;

    public function imgAlt(): string;

    public function imgSrc(): string;

    public function keywords(): array;

    public function links(): array;

    public function meta(): array;

    public function title(): string;

    public function url(): string;
}
