<?php

namespace Tests\Unit;

use App\Enums\FrequencyType;
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

    public function test_is_runnable(): void
    {
        $bot = Bot::factory()->news()->enabled()->makeOne(['frequency' => FrequencyType::Hourly, 'last_run_at' => now()->subHours(2)]);
        $this->assertTrue($bot->isRunnable());

        foreach ([FrequencyType::Daily, FrequencyType::Weekly, FrequencyType::Monthly] as $frequency) {
            $bot->frequency = $frequency;
            $this->assertFalse($bot->isRunnable());
        }

        $bot->frequency = FrequencyType::Hourly;
        $bot->enabled = false;
        $this->assertFalse($bot->isRunnable());
    }

    public function test_blacklisted(): void
    {
        $bot = Bot::factory()->news()->makeOne(['blacklist' => "foo\nbar"]);

        $this->assertFalse($bot->blacklisted());
        $this->assertTrue($bot->blacklisted('some foo thing'));
        $this->assertFalse($bot->blacklisted('nothing', 'here'));
        $this->assertSame([], Bot::factory()->makeOne(['blacklist' => '  '])->blacklist());
    }
}
