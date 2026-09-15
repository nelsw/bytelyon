<?php

namespace App\Actions;

use App\Concerns\SitemapValidationRules;
use App\Models\Bot;
use Illuminate\Support\Facades\Log;
use Illuminate\Support\Facades\Validator;

class UpdateSitemap
{
    use SitemapValidationRules;

    /** @param array<string, mixed> $urls */
    public function __invoke(Bot $bot, array $urls): void
    {
        ksort($urls);
        $attributes = Validator::validate([
            'urls' => array_keys($urls),
        ], $this->sitemapRules());

        $bot->sitemap?->update($attributes);

        Log::debug('Sitemap saved', [
            'bot' => $bot->id,
            'id' => $bot->sitemap?->id,
            'domain' => $bot->query,
            'urls' => count($attributes['urls']),
        ]);
    }
}
