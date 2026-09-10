<?php

namespace App\Jobs;

use App\Models\Bot;
use App\Services\BotService;
use Illuminate\Contracts\Queue\ShouldQueue;
use Illuminate\Foundation\Queue\Queueable;
use Illuminate\Queue\Attributes\Timeout;
use Illuminate\Queue\Middleware\WithoutOverlapping;
use Illuminate\Queue\SerializesModels;
use Illuminate\Support\Facades\Log;
use Throwable;

#[Timeout(60 * 5)]
class BotJob implements ShouldQueue
{
    use Queueable, SerializesModels;

    public function __construct(
        public readonly Bot $bot,
    ) {}

    public function middleware(): array
    {
        return [new WithoutOverlapping('bot')];
    }

    public function handle(BotService $service): void
    {
        if ($this->bot->isNotRunnable()) {
            return;
        }

        $service->run($this->bot);

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
