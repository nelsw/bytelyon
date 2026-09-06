<?php

namespace App\Services\Bots\News;

use App\Enums\BotType;
use App\Models\Bot;
use App\Services\Bots\News\Rss\BingRssService;
use App\Services\Bots\News\Rss\GoogleRssService;
use App\Services\Bots\ScriptsService;
use Illuminate\Container\Attributes\Singleton;
use Illuminate\Http\Client\ConnectionException;
use Illuminate\Http\Client\RequestException;
use Illuminate\Support\Facades\Log;

#[Singleton]
readonly class NewsBotService
{
    public function __construct(
        private GoogleRssService $googleRssService,
        private BingRssService   $bingRssService,
        private ScriptsService   $service,
    )
    {
    }

    public function run(Bot $bot): bool
    {
        $blacklist = explode("\n", $bot->blacklist);
        try {
            $arr = [
                ...$this->bingRssService->fetch($bot->query, $bot->played_at, $blacklist),
                ...$this->googleRssService->fetch($bot->query, $bot->played_at, $blacklist),
            ];
        } catch (ConnectionException|RequestException$e) {
            Log::error('NewsBotService:run', ['exception' => $e]);
            return false;
        }
        $urls = array_column($arr, 'url');

        if (!$this->service->data($bot->type, $bot->id, $urls, $bot->headless)) {
            return false;
        }

        // for each url,
        // get the data
        // set the data
        // save the article

        return true;
    }
}
