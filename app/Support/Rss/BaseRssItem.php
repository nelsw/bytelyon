<?php

namespace App\Support\Rss;

use App\Traits\HasNewsSource;
use Carbon\CarbonInterface;
use Illuminate\Support\Carbon;
use InvalidArgumentException;
use Log;
use SimpleXMLElement;

abstract class BaseRssItem extends SimpleXMLElement
{
    use HasNewsSource;

    abstract public function imageSrc(): string;

    abstract public function publisher(): string;

    public function description(): string
    {
        return (string) $this->description;
    }

    public function title(): string
    {
        return (string) $this->title;
    }

    public function url(): string
    {
        return (string) $this->link;
    }

    public function publishedAt(): CarbonInterface|string
    {
        return rescue(
            fn () => Carbon::parse((string) $this->pubDate),
            fn () => (string) $this->pubDate,
            fn (InvalidArgumentException $e) => Log::warning("RssItem#publishedAt: {$e->getMessage()}"));
    }

    /** @return array<string, string|CarbonInterface> */
    public function toArray(): array
    {
        $publishedAt = $this->publishedAt();
        return [
            'description' => $this->description(),
            'img_src' => $this->imageSrc(),
            'published_at' => is_string($publishedAt) ? $publishedAt : $publishedAt->toDateTimeString(),
            'publisher' => $this->publisher(),
            'source' => $this->source(),
            'title' => $this->title(),
            'url' => $this->url(),
        ];
    }
}
