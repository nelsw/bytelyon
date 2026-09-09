<?php

namespace App\Services\Bots\News;

use App\Dto\Meta;
use App\Models\Article;
use App\Models\Bot;
use App\Services\Bots\News\Rss\BingRssService;
use App\Services\Bots\News\Rss\GoogleRssService;
use Illuminate\Container\Attributes\Singleton;
use Illuminate\Http\Client\ConnectionException;
use Illuminate\Http\Client\RequestException;
use Illuminate\Support\Arr;
use Illuminate\Support\Facades\Log;
use Illuminate\Support\Facades\Storage;
use Illuminate\Support\Facades\URL;

#[Singleton]
final readonly class NewsBotService
{
    public function __construct(
        private GoogleRssService $googleRssService,
        private BingRssService   $bingRssService,
    ) {}

    public function run(Bot $bot): bool
    {
        try {
            /** @var array<Article> $arr */
            $arr = [
                ...$this->bingRssService->fetch($bot),
                ...$this->googleRssService->fetch($bot),
            ];
        } catch (ConnectionException|RequestException$e) {
            Log::error('NewsBotService', ['exception' => $e]);
            return false;
        }

        $arrCount = count($arr);
        Log::info('NewsBotService', [
            'bot' => $bot->toPrettyJson(),
            'articles' => $arrCount,
        ]);

        if ($arrCount === 0) {
            return true;
        } elseif (!(new Article)->process(array_column($arr, 'url'))) {
            return false;
        }

        Arr::map($arr, function (array $a) use ($bot) {
            $uuid = URL::toUuid5($a['url']);
            $data = Storage::json("{$bot->type->value}/$bot->id/$uuid.json");

            $meta = Meta::of($data['meta']);

            $a['keywords'] = $meta->keywords;
            $a['img_alt'] = $meta->imageAlt;
            if ($a['img_url'] === '') {
                $a['img_url'] = $meta->imageSrc;
            }
            if ($a['description'] === '') {
                $a['description'] = $meta->description;
            }
            $a['body'] = $data['body'];
            return $a;
        });

        return true;
    }
}
