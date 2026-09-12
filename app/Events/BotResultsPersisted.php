<?php

namespace App\Events;

use App\Models\Bot;
use Illuminate\Broadcasting\Channel;
use Illuminate\Broadcasting\InteractsWithSockets;
use Illuminate\Broadcasting\PrivateChannel;
use Illuminate\Contracts\Broadcasting\ShouldBroadcast;
use Illuminate\Foundation\Events\Dispatchable;
use Illuminate\Queue\SerializesModels;

class BotResultsPersisted implements ShouldBroadcast
{
    use Dispatchable, InteractsWithSockets, SerializesModels;

    public function __construct(
        public readonly Bot $bot,
        public readonly string $message,
    ) {}

    /** @return array<int, Channel> */
    public function broadcastOn(): array
    {
        return [
            new PrivateChannel($this->bot->user),
        ];
    }

    public function broadcastAs(): string
    {
        return 'bot.results.persisted';
    }

    /** @return array<string, mixed> */
    public function broadcastWith(): array
    {
        return [
            'toast' => [
                'type' => 'success',
                'message' => $this->message,
            ],
            'bot' => [
                'id' => $this->bot->id,
                'type' => $this->bot->type,
                'query' => $this->bot->query,
            ],
        ];
    }
}
