<?php

namespace Tests\Unit\Events;

use App\Events\BotResultsPersisted;
use App\Models\Bot;
use Illuminate\Broadcasting\PrivateChannel;
use Tests\TestCase;

class BotResultsPersistedTest extends TestCase
{
    public function test_broadcast_payload(): void
    {
        $bot = Bot::factory()->news()->createOneQuietly();
        $event = new BotResultsPersisted($bot, 'Done.');

        $this->assertSame('bot.results.persisted', $event->broadcastAs());

        [$channel] = $event->broadcastOn();
        $this->assertInstanceOf(PrivateChannel::class, $channel);
        $this->assertSame('private-'.$bot->user->broadcastChannel(), $channel->name);

        $this->assertSame([
            'toast' => ['type' => 'success', 'message' => 'Done.'],
            'bot' => ['id' => $bot->id, 'type' => $bot->type, 'query' => $bot->query],
        ], $event->broadcastWith());
    }
}
