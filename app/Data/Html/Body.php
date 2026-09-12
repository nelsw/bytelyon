<?php

namespace App\Data\Html;

use Dom\Element;
use Dom\HTMLDocument;
use Illuminate\Contracts\Support\Arrayable;

readonly class Body implements Arrayable
{
    public function __construct(public string $content){}

    public static function make(string $html): self {
        $doc = HTMLDocument::createFromString($html);
        foreach (['article', 'main', 'body', 'html'] as $selector) {

            $e = $doc->querySelector($selector);
            if ($e === null) {
                continue;
            }

            $txt = collect($e->querySelectorAll('p'))
                ->map(fn (Element $e) => trim($e->textContent ?? ''))
                ->reject(fn (Element $e) => empty($e->textContent))
                ->implode(" ");

            if (!empty($txt)) {
                return new self($txt);
            }
        }
        return new self('');
    }

    public function isEmpty(): bool
    {
        return empty($this->content);
    }

    public function toArray(): array
    {
        return ['content' => $this->content];
    }
}
