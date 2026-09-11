<?php

namespace App\Contracts;

use Carbon\CarbonInterface;
use Illuminate\Contracts\Support\Arrayable;

interface RssItem extends Arrayable
{
    public function description(): string;

    public function imageSrc(): string;

    public function publishedAt(): CarbonInterface|string;

    public function publisher(): string;

    public function source(): string;

    public function title(): string;

    public function url(): string;
}
