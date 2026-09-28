<?php

namespace Tests\Feature\Controller;

use App\Enums\BotType;
use App\Enums\FrequencyType;
use App\Models\Bot;
use App\Models\Proxy;
use App\Models\User;
use Inertia\Testing\AssertableInertia as Assert;
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

    public function test_index_filters_and_sorts(): void
    {
        $user = User::factory()->verified()->create();
        Bot::factory()->for($user)->news('alpha news')->enabled()->headless()->createOneQuietly();
        Bot::factory()->for($user)->news('beta news')->createOneQuietly(['enabled' => false, 'headless' => false]);

        $this->actingAs($user)
            ->get(route('bots.index', ['query' => 'ALPHA', 'type' => 'news', 'status' => 'enabled', 'mode' => 'headless', 'sort' => 'query_desc', 'perPage' => 25]))
            ->assertOk()
            ->assertInertia(fn (Assert $page) => $page
                ->component('Dashboard')
                ->has('bots.data', 1)
                ->where('bots.data.0.query', 'alpha news')
                ->where('filters.perPage', 25)
            );

        $this->actingAs($user)
            ->get(route('bots.index', ['status' => 'disabled', 'mode' => 'headed']))
            ->assertInertia(fn (Assert $page) => $page
                ->has('bots.data', 1)
                ->where('bots.data.0.query', 'beta news')
            );
    }

    public function test_index_falls_back_to_default_sort_and_page_size(): void
    {
        $user = User::factory()->verified()->create();

        $this->actingAs($user)
            ->get(route('bots.index', ['sort' => 'bogus', 'perPage' => 7]))
            ->assertInertia(fn (Assert $page) => $page
                ->where('filters.sort', 'created_at_desc')
                ->where('filters.perPage', 10)
            );
    }

    public function test_index_accepts_every_sort(): void
    {
        $user = User::factory()->verified()->create();
        Bot::factory()->for($user)->news()->count(2)->createQuietly();

        foreach (['created_at_asc', 'query_asc', 'query_desc', 'type_asc', 'type_desc', 'enabled_desc', 'enabled_asc', 'headless_desc', 'headless_asc', 'processed_at_desc', 'processed_at_asc', 'updated_at_desc', 'updated_at_asc'] as $sort) {
            $this->actingAs($user)
                ->get(route('bots.index', ['sort' => $sort]))
                ->assertOk()
                ->assertInertia(fn (Assert $page) => $page->where('filters.sort', $sort)->has('bots.data', 2));
        }
    }

    public function test_show_create_and_edit(): void
    {
        $bot = Bot::factory()->news()->createOneQuietly();
        Proxy::factory()->create(['user_id' => $bot->user_id]);

        $this->actingAs($bot->user)
            ->get(route('bots.show', $bot))
            ->assertInertia(fn (Assert $page) => $page->component('bots/Show')->where('bot.id', $bot->id));

        $this->actingAs($bot->user)
            ->get(route('bots.create'))
            ->assertInertia(fn (Assert $page) => $page->component('bots/Create')->has('typeOptions')->has('frequencyOptions')->has('proxyOptions', 1));

        $this->actingAs($bot->user)
            ->get(route('bots.edit', $bot))
            ->assertInertia(fn (Assert $page) => $page->component('bots/Edit')->where('bot.id', $bot->id)->has('proxyOptions', 1));
    }

    public function test_other_users_cannot_view_edit_or_delete(): void
    {
        $bot = Bot::factory()->news()->createOneQuietly();
        $stranger = User::factory()->verified()->create();

        $this->actingAs($stranger)->get(route('bots.show', $bot))->assertForbidden();
        $this->actingAs($stranger)->get(route('bots.edit', $bot))->assertForbidden();
        $this->actingAs($stranger)->delete(route('bots.destroy', $bot))->assertForbidden();
    }

    public function test_destroy(): void
    {
        $bot = Bot::factory()->news()->createOneQuietly();

        $this->actingAs($bot->user)
            ->delete(route('bots.destroy', $bot))
            ->assertRedirect(route('dashboard'));

        $this->assertModelMissing($bot);
    }
}
