<?php

namespace App\Actions;

use App\Models\Sitemap;
use Illuminate\Support\Facades\Log;

class UpdateSitemapUrls
{

    public function __invoke(Sitemap $sitemap): void
    {
        ksort($sitemap->urls);

        $sitemap->save();

        Log::debug('Sitemap saved', [
            'bot' => $sitemap->bot_id,
            'id' => $sitemap->id,
            'domain' => $sitemap->domain,
            'urls' => count($sitemap->urls),
        ]);
    }
}
