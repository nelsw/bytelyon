<?php

namespace Tests\Feature\Service;

use App\Models\Article;
use App\Models\Bot;
use App\Services\LambdaService;
use App\Services\NewsBotService;
use App\Services\PageService;
use Illuminate\Support\Facades\App;
use Tests\TestCase;

class NewsBotServiceTest extends TestCase
{
    private NewsBotService $service;

    protected function setUp(): void
    {
        parent::setUp();
        $this->service = resolve(NewsBotService::class, [
            'pageService' => resolve(PageService::class),
        ]);
    }

    public function test_run(): void
    {
        if (! App::hasDebugModeEnabled()) {
            return;
        }

        $bot = Bot::factory()
            ->news()
            ->enabled()
            ->query('eth forecast')
            ->lastRunAt(now()->subWeek())
            ->createOneQuietly();

        $bot->user->proxies()->create(config('lambda.proxy'));

        $startedAt = now();
        $this->assertDoesntThrow(fn () => $this->service->run($bot));

        dump(Article::query()
            ->whereBotId($bot->id)
            ->get()
            ->toPrettyJson());
    }
}
