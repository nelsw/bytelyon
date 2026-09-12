<?php

namespace App\Jobs;

use App\Enums\BotType;
use App\Models\Bot;
use App\Services\NewsBotService;
use App\Services\SearchBotService;
use App\Services\SitemapBotService;
use Illuminate\Contracts\Queue\ShouldBeUnique;
use Illuminate\Contracts\Queue\ShouldQueue;
use Illuminate\Foundation\Queue\Queueable;
use Illuminate\Queue\Attributes\Timeout;
use Illuminate\Queue\SerializesModels;
use Illuminate\Support\Facades\Log;
use Throwable;

#[Timeout(60 * 5)]
class BotJob implements ShouldBeUnique, ShouldQueue
{
    use Queueable, SerializesModels;

    public function __construct(
        public readonly Bot $bot,
    ) {}

    public function uniqueId(): string
    {
        return strval($this->bot->id);
    }

    public function handle(
        NewsBotService $newsBotService,
        SearchBotService $searchBotService,
        SitemapBotService $sitemapBotService,
    ): void
    {
        switch ($this->bot->type) {
            case BotType::News:
                $newsBotService->run($this->bot);
                break;
            case BotType::Search:
                $searchBotService->run($this->bot);
                break;
            case BotType::Sitemap:
                $sitemapBotService->run($this->bot);
                break;
        }

        $this->bot->update(['last_run_at' => now()->utc()]);

        Log::info('BotJob::handle - worked', [
            'id' => $this->bot->id,
            'type' => $this->bot->type,
            'query' => $this->bot->query,
        ]);
    }

    public function failed(?Throwable $e): void
    {
        Log::error('BotJob::failed', [
            'exception' => $e,
            'id' => $this->bot->id,
            'type' => $this->bot->type,
            'query' => $this->bot->query,
        ]);
    }
}
