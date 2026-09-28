<?php

namespace Tests\Feature\Controller;

use App\Models\Article;
use App\Models\Bot;
use App\Models\User;
use Inertia\Testing\AssertableInertia as Assert;
use Tests\TestCase;

class NewsControllerTest extends TestCase
{
    public function test_index_lists_the_users_news_bots(): void
    {
        $user = User::factory()->verified()->create();
        $bot = Bot::factory()->for($user)->news('zeta')->createOneQuietly();
        Article::factory()->for($bot)->count(2)->createQuietly();
        Bot::factory()->for($user)->news('alpha')->createOneQuietly();
        Bot::factory()->news()->createOneQuietly();

        $this->actingAs($user)
            ->get(route('news.index'))
            ->assertOk()
            ->assertInertia(fn (Assert $page) => $page
                ->component('news/Index')
                ->has('bots', 2)
                ->where('bots.0.query', 'alpha')
                ->where('bots.1.query', 'zeta')
            );
    }
}
