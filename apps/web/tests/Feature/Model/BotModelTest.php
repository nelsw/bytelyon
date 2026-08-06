<?php

namespace Tests\Feature\Model;

use App\Models\Bot;
use App\Models\User;
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
        Bot::query()->forceDelete();

        $exp = Bot::factory()
            ->enabled()
            ->neverRun()
            ->create();

        $act = Bot::query()
            ->enabled()
            ->ready()
            ->get()
            ->first();

        $this->assertEquals($exp->id, $act->id);
    }
}
