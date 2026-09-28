<?php

namespace App\Jobs;

use App\Facades\Go;
use App\Models\Bot;
use Illuminate\Bus\Batchable;
use Illuminate\Contracts\Queue\ShouldBeUnique;
use Illuminate\Contracts\Queue\ShouldQueue;
use Illuminate\Foundation\Queue\Queueable;
use Illuminate\Queue\SerializesModels;
use Illuminate\Support\Facades\Log;
use Throwable;

class BotJob implements ShouldBeUnique, ShouldQueue
{
    use Batchable, Queueable, SerializesModels;

    public function __construct(
        public readonly Bot $bot,
    ) {}

    public function uniqueId(): string
    {
        return "bot:{$this->bot->id}";
    }

    public function prepareForDispatch(): bool
    {
        return $this->bot->isRunnable();
    }

    public function handle(): void
    {
        Log::debug('BotJob');
        Go::pub($this->bot);
    }

    public function failed(?Throwable $e): void
    {
        Log::error('BotJob', ['exception' => $e]);
    }
}
