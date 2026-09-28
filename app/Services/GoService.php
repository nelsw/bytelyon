<?php

namespace App\Services;

use App\Events\BotResultsPersisted;
use App\Models\Bot;
use Closure;
use Illuminate\Database\Eloquent\ModelNotFoundException;
use Illuminate\Redis\Connections\Connection;
use Illuminate\Support\Facades\Log;
use Illuminate\Support\Facades\Redis;
use Illuminate\Support\Fluent;

class GoService
{
    protected Connection $client;

    public function __construct()
    {
        $this->client = Redis::connection('go');
    }

    public function pub(Bot|int $bot): void {
        if (is_int($bot)) {
            try {
                $bot = Bot::query()->findOrFail($bot);
            } catch (ModelNotFoundException $e) {
                return;
            }
        }

        $this->client->publish('bots', json_encode([
            'id' => $bot->id,
            'child_id' => $bot->childId(),
            'type' => $bot->type->value,
            'headless' => $bot->headless,
            'query' => $bot->query,
            'last_run_at' => $bot->lastRunAt(),
            'blacklist' => $bot->blacklist,
        ]));
    }

    public function sub(Closure $callback): void {
        $this->client->subscribe('evts', function (string $message) use ($callback) {
            $callback($message);
            $ƒ = Fluent::of($message);
            try {
                BotResultsPersisted::dispatch(
                    Bot::query()->findOrFail($ƒ->integer('id')),
                    $ƒ->string('message'),
                );
            } catch (ModelNotFoundException $e) {
                Log::error('GoService#subscribe', [
                    'channel' => 'evts',
                    'message' => $message,
                    'exception' => $e,
                ]);
            }
        });
    }
}
