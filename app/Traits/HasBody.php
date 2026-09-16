<?php

namespace App\Traits;

use Dom\Element;

trait HasBody
{
    public function body(): string
    {
        foreach (['article', 'main', 'body', 'html'] as $tag) {
            if ($this->isMissing($tag)) {
                continue;
            }

            $values = collect($this->element($tag)->querySelectorAll('p'))
                ->transform(fn (Element $e): string => trim($e->textContent ?? ''))
                ->reject(fn (string $href): bool => empty($href));

            if ($values->isNotEmpty()) {
                return $values->join(' ');
            }
        }

        return '';
    }
}
