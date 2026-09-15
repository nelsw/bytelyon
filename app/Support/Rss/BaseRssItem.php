<?php

namespace App\Support\Rss;

use App\Contracts\RssItem;
use Carbon\CarbonInterface;
use Carbon\Exceptions\InvalidArgumentException;
use Illuminate\Support\Carbon;
use Illuminate\Support\Facades\Log;
use SimpleXMLElement;

abstract class BaseRssItem extends SimpleXMLElement implements RssItem
{
    abstract public function imageSrc(): string;

    abstract public function publisher(): string;

    abstract public function source(): string;

    public function publishedAt(): CarbonInterface|string
    {
        try {
            return Carbon::parse((string) $this->pubDate);
        } catch (InvalidArgumentException $e) {
            Log::warning("RssItem#publishedAt: {$e->getMessage()}");
            return (string) $this->pubDate;
        }
    }

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

    public function toArray(): array
    {
        return [
            'description' => $this->description(),
            'img_src' => $this->imageSrc(),
            'published_at' => $this->publishedAt()->toDateTimeString(),
            'publisher' => $this->publisher(),
            'source' => $this->source(),
            'title' => $this->title(),
            'url' => $this->url(),
        ];
    }
}
