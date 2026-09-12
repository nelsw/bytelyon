<?php

namespace App\Data;

use Closure;

readonly class Entry
{

    public function __construct(
        public ?string $key,
        public ?string $val,
    )
    {
    }

    public function isKeyEmpty(): bool
    {
        return empty($this->key);
    }

    public function isValEmpty(): bool
    {
        return empty($this->val);
    }

    public function assert(Closure $validator): bool
    {
        return $validator($this->key, $this->val);
    }

    public function toArray(): array
    {
        return [
            'key' => $this->key,
            'val' => $this->val
        ];
    }

    public function toItem(): array
    {
        return [$this->key => $this->val];
    }
}
