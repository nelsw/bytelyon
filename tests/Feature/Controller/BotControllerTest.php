<?php

namespace Tests\Feature\Controller;

use App\Enums\BotType;
use App\Enums\FrequencyType;
use App\Models\Bot;
use App\Models\User;
use Tests\TestCase;

class BotControllerTest extends TestCase
{
    public function test_store_rejects_a_duplicate_query_for_the_same_type(): void
    {
        $user = User::factory()->verified()->create();

        Bot::factory()->for($user)->create([
            'type' => BotType::News,
            'query' => 'laravel release notes',
        ]);

        $response = $this->actingAs($user)->post(route('bots.store'), [
            'blacklist' => null,
            'enabled' => true,
            'headless' => true,
            'frequency' => FrequencyType::values()[0],
            'type' => BotType::News->value,
            'query' => 'laravel release notes',
        ]);

        $response->assertSessionHasErrors('query');
        $this->assertSame(1, Bot::where('user_id', $user->id)->count());
    }

    public function test_store_allows_the_same_query_for_a_different_type(): void
    {
        $user = User::factory()->verified()->create();

        Bot::factory()->for($user)->create([
            'type' => BotType::News,
            'query' => 'laravel release notes',
        ]);

        $response = $this->actingAs($user)->post(route('bots.store'), [
            'blacklist' => null,
            'enabled' => true,
            'headless' => true,
            'frequency' => FrequencyType::values()[0],
            'type' => BotType::Search->value,
            'query' => 'laravel release notes',
        ]);

        $response->assertSessionDoesntHaveErrors();
        $this->assertSame(2, Bot::where('user_id', $user->id)->count());
    }
}
