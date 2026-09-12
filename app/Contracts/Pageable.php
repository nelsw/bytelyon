<?php

namespace App\Contracts;


use Illuminate\Contracts\Support\Arrayable;

interface Pageable extends Arrayable
{
    public function url(): string;
    public function description(): string;

    public function keywords(): array;

    public function imgSrc(): string;
    public function imgAlt(): string;
    public function domain(): string;
    public function title(): string;
    public function body(): string;
    public function links(): array;
}
