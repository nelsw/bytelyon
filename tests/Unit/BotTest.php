<?php

namespace Tests\Unit;

use App\Models\Bot;
use App\Models\Proxy;
use Tests\TestCase;

class BotTest extends TestCase
{
    public function test_random_proxy_returns_null_when_none_are_attached(): void
    {
        $bot = Bot::factory()->news()->createOneQuietly();

        $this->assertNull($bot->randomProxy());
    }

    public function test_random_proxy_returns_the_only_attached_proxy(): void
    {
        $bot = Bot::factory()->news()->createOneQuietly();
        $proxy = Proxy::factory()->create(['user_id' => $bot->user_id]);
        $bot->proxies()->attach($proxy);

        $this->assertTrue($bot->randomProxy()->is($proxy));
    }

    public function test_random_proxy_returns_one_of_several_attached_proxies(): void
    {
        $bot = Bot::factory()->news()->createOneQuietly();
        $proxies = Proxy::factory(5)->create(['user_id' => $bot->user_id]);
        $bot->proxies()->attach($proxies);

        $this->assertContains($bot->randomProxy()->id, $proxies->pluck('id')->all());
    }
}
