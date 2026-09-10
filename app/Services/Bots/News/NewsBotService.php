<?php

namespace App\Services\Bots\News;

use App\Dto\Meta;
use App\Models\Bot;
use App\Services\BotProcessService;
use App\Services\Bots\News\Rss\BingRssService;
use App\Services\Bots\News\Rss\GoogleRssService;
use Illuminate\Container\Attributes\Singleton;
use Illuminate\Support\Facades\Log;
use Illuminate\Support\Facades\Storage;
use Illuminate\Support\Facades\URL;

#[Singleton]
final readonly class NewsBotService
{
    public function __construct(
        private GoogleRssService $googleRssService,
        private BingRssService $bingRssService,
        private BotProcessService $botProcessService,
    ) {}

    public function run(Bot $bot): bool
    {
        $articles = collect()
            ->merge($this->bingRssService->fetch($bot))
            ->merge($this->googleRssService->fetch($bot));

        Log::info('NewsBotService', [
            'bot' => $bot->toPrettyJson(),
            'articles' => $articles->count(),
        ]);

        if ($articles->isEmpty()) {
            return true;
        }

        $articles->each(function (array $arr) use ($bot) {

            $url = $arr['url'];
            if (! $this->botProcessService->url($bot, $url)) {
                return;
            }

            $uuid = URL::toUuid5($url);
            $data = Storage::json("{$bot->type->value}/$bot->id/$uuid.json");
            if ($data === null) {
                return;
            }
            $meta = Meta::of($data['meta']);
            $body = $data['body'];

            $bot->articles()->updateOrCreate(
                ['url' => $url],
                [
                    ...$arr,
                    ...[
                        'keywords' => $meta->keywords,
                        'img_alt' => $meta->imageAlt,
                        'img_url' => $meta->imageSrc,
                        'description' => $meta->description,
                        'body' => $body,
                    ],
                ],
            );
        });

        return true;
    }
}
