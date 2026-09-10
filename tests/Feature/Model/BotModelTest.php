<?php

namespace Tests\Feature\Model;

use App\Models\Bot;
use App\Models\User;
use Illuminate\Support\Facades\Process;
use Tests\TestCase;

class BotModelTest extends TestCase
{
    public function test_it(): void
    {
        /** @var Bot $bot */
        $bot = Bot::factory()
            ->for(User::factory())
            ->enabled()
            ->neverRun()
            ->create();

        $this->assertInstanceOf(Bot::class, $bot);
        $this->assertDatabaseHas('bots', $bot->attributesToArray());
    }

    public function test_ready(): void
    {
        $result = Process::run('ls -la');
        dump($result->output());

        Bot::query()->forceDelete();

        /** @var Bot $exp */
        $exp = Bot::factory()
            ->enabled()
            ->neverRun()
            ->create([
                'headless' => false,
                'query' => 'btc forecast',
            ]);

        /** @var Bot $act */
        $act = Bot::query()
            ->enabled()
            ->ready()
            ->get()
            ->first();

        $this->assertEquals($exp->id, $act->id);
        dump($act->toJson());
        $result = Process::run(base_path('scripts/main')." '".$act->toJson()."'");
        dump($result);
        dump($result->output());
    }
}
