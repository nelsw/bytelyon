<?php

namespace Tests\Feature\Controller;

use App\Enums\BotType;
use App\Enums\FrequencyType;
use App\Models\Bot;
use App\Models\Proxy;
use App\Models\User;
use Tests\TestCase;

class BotControllerTest extends TestCase
{
    public function test_store_rejects_a_duplicate_query_for_the_same_type(): void
    {
        $bot = Bot::factory()->news()->createOneQuietly();

        $response = $this->actingAs($bot->user)->post(route('bots.store'), [
            'blacklist' => null,
            'enabled' => true,
            'headless' => true,
            'frequency' => FrequencyType::values()[0],
            'type' => $bot->type->value,
            'query' => $bot->query,
        ]);

        $response->assertSessionHasErrors('query');
        $this->assertSame(1, Bot::where('user_id', $bot->user_id)->count());
    }

    public function test_store_allows_the_same_query_for_a_different_type(): void
    {
        $bot = Bot::factory()->news()->createOneQuietly();

        $response = $this->actingAs($bot->user)->post(route('bots.store'), [
            'blacklist' => null,
            'enabled' => true,
            'headless' => true,
            'frequency' => FrequencyType::values()[0],
            'type' => BotType::Search->value,
            'query' => $bot->query,
        ]);

        $response->assertSessionDoesntHaveErrors();
        $this->assertSame(2, Bot::where('user_id', $bot->user_id)->count());
    }

    public function test_store_attaches_the_selected_proxies(): void
    {
        $user = User::factory()->verified()->create();
        $proxies = Proxy::factory(2)->create(['user_id' => $user->id]);

        $response = $this->actingAs($user)->post(route('bots.store'), [
            'blacklist' => null,
            'enabled' => true,
            'headless' => true,
            'frequency' => FrequencyType::values()[0],
            'type' => BotType::News->value,
            'query' => 'a valid query',
            'proxies' => $proxies->pluck('id')->all(),
        ]);

        $response->assertSessionDoesntHaveErrors();

        $bot = Bot::where('user_id', $user->id)->sole();
        $this->assertSame($proxies->pluck('id')->sort()->values()->all(), $bot->proxies->pluck('id')->sort()->values()->all());
    }

    public function test_store_rejects_a_proxy_belonging_to_another_user(): void
    {
        $user = User::factory()->verified()->create();
        $otherUsersProxy = Proxy::factory()->create();

        $response = $this->actingAs($user)->post(route('bots.store'), [
            'blacklist' => null,
            'enabled' => true,
            'headless' => true,
            'frequency' => FrequencyType::values()[0],
            'type' => BotType::News->value,
            'query' => 'a valid query',
            'proxies' => [$otherUsersProxy->id],
        ]);

        $response->assertSessionHasErrors('proxies.0');
        $this->assertSame(0, Bot::where('user_id', $user->id)->count());
    }

    public function test_update_syncs_the_selected_proxies(): void
    {
        $bot = Bot::factory()->news()->createOneQuietly();
        $originalProxy = Proxy::factory()->create(['user_id' => $bot->user_id]);
        $bot->proxies()->attach($originalProxy);

        $replacementProxy = Proxy::factory()->create(['user_id' => $bot->user_id]);

        $response = $this->actingAs($bot->user)->put(route('bots.update', $bot), [
            'blacklist' => null,
            'enabled' => true,
            'headless' => true,
            'frequency' => FrequencyType::values()[0],
            'proxies' => [$replacementProxy->id],
        ]);

        $response->assertSessionDoesntHaveErrors();
        $bot->refresh();
        $this->assertSame([$replacementProxy->id], $bot->proxies->pluck('id')->all());
    }

    public function test_update_with_no_proxies_selected_detaches_all_proxies(): void
    {
        $bot = Bot::factory()->news()->createOneQuietly();
        $bot->proxies()->attach(Proxy::factory()->create(['user_id' => $bot->user_id]));

        $response = $this->actingAs($bot->user)->put(route('bots.update', $bot), [
            'blacklist' => null,
            'enabled' => true,
            'headless' => true,
            'frequency' => FrequencyType::values()[0],
        ]);

        $response->assertSessionDoesntHaveErrors();
        $this->assertSame(0, $bot->proxies()->count());
    }
}
