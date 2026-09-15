<?php

namespace App\Traits;

use Dom\Element;

trait HasElement
{
    protected function hasElement(): void
    {
        $this->emptyElement = resolve(Element::class);
        $this->emptyElement->rename(null, 'null');
    }
}
