<?php

namespace App\Jobs;

use App\Models\Bot;
use Illuminate\Contracts\Queue\ShouldBeUnique;
use Illuminate\Contracts\Queue\ShouldQueue;
use Illuminate\Foundation\Queue\Queueable;
use Illuminate\Queue\SerializesModels;
use Illuminate\Support\Facades\Log;
use Throwable;

class BaseBotJob implements ShouldBeUnique, ShouldQueue
{
    use Queueable, SerializesModels;

    public int $timeout = 0;
    public function __construct(
        public readonly Bot $bot,
    ) {
        $this->onConnection('sqs');

    }

    public function uniqueId(): string
    {
        return strval($this->bot->id);
    }

    public function messageGroup(): string
    {
        return 'bots';
    }

    public function prepareForDispatch(): bool
    {
        return $this->bot->isRunnable();
    }

    public function failed(?Throwable $e): void
    {
        $this->bot->update(['last_run_at' => now()->utc()]);
        Log::error('BotJob', [
            'exception' => $e,
            'id' => $this->bot->id,
            'type' => $this->bot->type->value,
            'query' => $this->bot->query,
        ]);
    }
}
