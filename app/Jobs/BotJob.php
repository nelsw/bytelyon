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

        if (! $this->check()) {
            return;
        }

        $payload = $this->prepare();

        $result = $this->run($payload);

        Log::info('BotJob::handle - worked', [
            ...$payload,
            ...['result' => $result],
        ]);

        if ($result !== 'ok') {
            $this->fail($result);
        }
    }

    private function check(): bool
    {
        if (! $this->bot->enabled) {
            Log::info('BotJob::handle - bot disabled', ['id' => $this->bot->id]);
            return false;
        }

        if ($this->bot->last_run_at === null) {
            return true;
        }

        $nextRunAt = $this->bot->last_run_at->add($this->bot->frequency->interval());
        if ($nextRunAt->isFuture()) {
            Log::info('BotJob::handle - too soon', [
                'id' => $this->bot->id,
                'nextRunAt' => $nextRunAt,
            ]);
            return false;
        }

        return true;
    }

    private function prepare(): array
    {
        $payload = [
            'id' => $this->bot->id,
            'type' => $this->bot->type,
            'query' => $this->bot->query,
            'headless' => $this->bot->headless,
            'last_run_at' => ($this->bot->last_run_at ?? now()->subYear()),
        ];

        if ($this->bot->type === BotType::Sitemap) {
            $payload['sitemap_id'] = $this->bot->sitemap()->firstOrCreate(['domain' => $payload['query']])->id;
        } elseif ($this->bot->type === BotType::Search) {
            $payload['serp_id'] = $this->bot->serp()->firstOrCreate(['query' => $payload['query']])->id;
        }

        return $payload;
    }

    private function run(array $payload): string
    {
        $bro = Redis::connection('broker');
        $bro->del("bot:{$this->bot->id}:done");
        $bro->set("bot:{$this->bot->id}:ready", json_encode($payload));
        while ($bro->get("bot:{$this->bot->id}:done") === null) {
            if ($bro->exists("bot:{$this->bot->id}:ready")) {
                Log::debug('BotJob::handle - waiting...');
                sleep(30);
            } else {
                Log::debug('BotJob::handle - working!!!');
                sleep(60);
            }
        }
        return $bro->getDel("bot:{$this->bot->id}:done");
    }

    public function failed(?Throwable $e): void
    {
        Log::error('BotJob::failed', [
            'exception' => $e,
            'bot.id' => $this->uniqueId(),
        ]);
    }
}
