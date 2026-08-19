<?php

namespace App\Jobs;

use App\Enums\BotType;
use App\Models\Bot;
use Illuminate\Contracts\Queue\ShouldBeUnique;
use Illuminate\Contracts\Queue\ShouldQueue;
use Illuminate\Foundation\Queue\Queueable;
use Illuminate\Queue\Attributes\Timeout;
use Illuminate\Queue\SerializesModels;
use Illuminate\Support\Facades\Log;
use Illuminate\Support\Facades\Redis;
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

    public function handle(): void
    {

        if ($this->bot->isNotRunnable()) {
            return;
        }

        $bro = Redis::connection('broker');

        $bro->del("bot:{$this->bot->id}:done");
        $bro->set("bot:{$this->bot->id}:ready", $this->bot->toJson());

        while ($bro->get("bot:{$this->bot->id}:done") === null) {
            if ($bro->exists("bot:{$this->bot->id}:ready")) {
                Log::debug('BotJob::handle - waiting...');
                sleep(30);
            } else {
                Log::debug('BotJob::handle - working!!!');
                sleep(60);
            }
        }

        $result = $bro->getDel("bot:{$this->bot->id}:done");

        Log::info('BotJob::handle - worked', [
            'type' => $this->bot->type,
            'query' => $this->bot->query,
            'result' => $result
        ]);

        if ($result !== 'ok') {
            $this->fail($result);
        }
    }

    public function failed(?Throwable $e): void
    {
        Log::error('BotJob::failed', [
            'exception' => $e,
            'bot.id' => $this->uniqueId(),
        ]);
    }
}
