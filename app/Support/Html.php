<?php

namespace App\Support;

use Dom\HTMLDocument;
use Illuminate\Support\Uri;

readonly class Html
{
    final public function __construct(
        public string       $url,
        public string       $domain,
        private HTMLDocument $doc,
    ){}

    public static function of(string $url, ?string $content = null): static
    {
        return new static(Uri::clean($url), Uri::domain($url), rescue(
            fn() => HTMLDocument::createFromString($content, LIBXML_NOERROR | LIBXML_HTML_NOIMPLIED),
            HTMLDocument::createEmpty()));
    }


    /** @return array<string, bool> */
    public function links(): array
    {
        /** @var array<string, bool> $links */
        $links = [];
        foreach ($this->doc->querySelectorAll('a') as $e) {
            $href = str($e->getAttribute('href'))->trim()->rtrim('/');
            if ($href->startsWith(["https://$this->domain", "https://www.$this->domain"])) {
                $links[$href->toString()] = false;
            } elseif ($href->startsWith('/')) {
                $uri = Uri::of($this->url);
                $links[$href->prepend($uri->scheme(), '://', $uri->host())->toString()] = false;
            }
        }
        return $links;
    }

    /** @return array<string, string> */
    public function meta(): array
    {
        $meta = [];
        foreach ($this->doc->querySelectorAll('meta') as $e) {
            $val = trim($e->getAttribute('content'));
            if (empty($val)) {
                continue;
            }

            $key = trim($e->getAttribute('name'));
            if (empty($key)) {
                $key = trim($e->getAttribute('property'));
                if (empty($key)) {
                    continue;
                }
            }

            if (!isset($meta[$key])) {
                $meta[$key] = $val;
            }
        }
        return $meta;
    }

    public function title(): string
    {
        return $this->doc->title;
    }
}
