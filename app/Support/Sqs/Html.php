<?php

namespace App\Support\Sqs;

use App\Traits\HasBody;
use App\Traits\HasLinks;
use App\Traits\HasMeta;
use Dom\Element;
use Dom\HTMLDocument;
use Dom\NodeList;
use Illuminate\Support\Str;
use Illuminate\Support\Stringable;
use IteratorAggregate;

abstract class Html
{
    use HasBody, HasMeta;

    protected HTMLDocument $document;

    protected Element $emptyElement;

    public function __construct(
        protected ?string $content = null,
    ) {
        $this->document = rescue(
            fn () => HTMLDocument::createFromString(
                source: $this->content ?? '',
                options: LIBXML_NOERROR | LIBXML_HTML_NOIMPLIED,
            ), HTMLDocument::createEmpty(),
        );
        $this->emptyElement = $this->document->createElement('div');
    }

    public function title(): string
    {
        return $this->document->title;
    }

    public function attr(string $selector, string $name): Stringable
    {
        return Str::of($this->element($selector)->getAttribute($name));
    }

    public function element(string $selector): Element
    {
        return $this->document->querySelector($selector) ?? $this->emptyElement;
    }

    /**
     * @implements IteratorAggregate<int, Element>
     *
     * @return NodeList<Element>
     */
    public function elements(string $selector): NodeList
    {
        return $this->document->querySelectorAll($selector);
    }

    public function isMissing(string $selector): bool
    {
        return $this->document->querySelector($selector) === null;
    }
}
