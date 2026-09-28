<?php

namespace Tests\Feature\Services;

use App\Events\BotResultsPersisted;
use App\Models\Bot;
use App\Services\GoService;
use Illuminate\Redis\Connections\Connection;
use Illuminate\Support\Facades\Event;
use Illuminate\Support\Facades\Log;
use Illuminate\Support\Facades\Redis;
use Mockery;
use Mockery\MockInterface;
use Tests\TestCase;

class GoServiceTest extends TestCase
{
    private MockInterface $connection;

    protected function setUp(): void
    {
        parent::setUp();

        $this->connection = Mockery::mock(Connection::class);
        Redis::shouldReceive('connection')->with('go')->andReturn($this->connection);
    }

    public function test_pub_publishes_the_bot(): void
    {
        $bot = Bot::factory()->news()->neverRun()->createOneQuietly();

        $this->connection->shouldReceive('publish')->once()->withArgs(function (string $channel, string $payload) use ($bot) {
            $data = json_decode($payload, true);

            return $channel === 'bots'
                && $data['id'] === $bot->id
                && $data['child_id'] === -1
                && $data['type'] === 'news'
                && $data['query'] === $bot->query;
        });

        new GoService()->pub($bot->id);
    }

    public function test_pub_ignores_a_missing_bot(): void
    {
        $this->connection->shouldNotReceive('publish');

        new GoService()->pub(PHP_INT_MAX);
    }

    public function test_sub_dispatches_results_for_known_bots(): void
    {
        Event::fake([BotResultsPersisted::class]);
        $bot = Bot::factory()->news()->createOneQuietly();
        $message = json_encode(['id' => $bot->id, 'message' => 'Found 3 articles.']);

        $this->connection->shouldReceive('subscribe')->once()
            ->andReturnUsing(fn (string $channel, callable $handler) => $handler($message));

        $received = null;
        new GoService()->sub(function (string $m) use (&$received) {
            $received = $m;
        });

        $this->assertSame($message, $received);
        Event::assertDispatched(BotResultsPersisted::class, fn (BotResultsPersisted $e) => $e->bot->is($bot) && $e->message === 'Found 3 articles.');
    }

    public function test_sub_logs_unknown_bots(): void
    {
        Event::fake([BotResultsPersisted::class]);
        Log::spy();

        $this->connection->shouldReceive('subscribe')->once()
            ->andReturnUsing(fn (string $channel, callable $handler) => $handler(json_encode(['id' => PHP_INT_MAX, 'message' => 'x'])));

        new GoService()->sub(fn () => null);

        Event::assertNotDispatched(BotResultsPersisted::class);
        Log::shouldHaveReceived('error')->once()->withArgs(fn (string $msg) => $msg === 'GoService#subscribe');
    }
}
