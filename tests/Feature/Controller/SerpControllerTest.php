<?php

namespace Tests\Feature\Controller;

use App\Facades\S3;
use App\Models\Serp;
use App\Models\User;
use Inertia\Testing\AssertableInertia as Assert;
use Tests\TestCase;

class SerpControllerTest extends TestCase
{
    public function test_authenticated_verified_users_can_view_available_serps(): void
    {
        Serp::query()->delete();

        $user = User::factory()->verified()->create();

        Serp::factory()->createQuietly();

        Serp::factory()->deleted()->createQuietly();

        $response = $this->actingAs($user)->get(route('serps.index'));

        $response->assertOk();

        $response->assertInertia(fn (Assert $page) => $page
            ->component('serps/Index')
            ->has('serps', 1)
        );
    }

    public function test_show_maps_results(): void
    {
        S3::shouldReceive('url')->andReturn('https://signed.example.com/shot.png');
        $user = User::factory()->verified()->create();
        $serp = Serp::factory()->createQuietly([
            'data' => [
                'similar_queries' => [['value' => 'foo'], ['value' => null], ['value' => 'bar']],
                'organic_results' => [
                    ['index' => 1, 'title' => 'Laravel', 'url' => 'https://www.laravel.com/docs', 'domain' => 'laravel.com', 'image' => 'https://img', 'snippet' => 'Docs'],
                    ['title' => 'No url'],
                ],
            ],
        ]);

        $this->actingAs($user)
            ->get(route('serps.show', $serp))
            ->assertOk()
            ->assertInertia(fn (Assert $page) => $page
                ->component('serps/Show')
                ->where('serp.similarQueries', ['foo', 'bar'])
                ->where('serp.screenshotUrl', 'https://signed.example.com/shot.png')
                ->has('serp.results', 2)
                ->where('serp.results.0.id', 'organic_results-1')
                ->where('serp.results.0.kind', 'Organic Results')
                ->where('serp.results.0.faviconUrl', 'https://www.google.com/s2/favicons?domain=laravel.com&sz=32')
                ->where('serp.results.0.imageUrl', 'https://img')
                ->where('serp.results.0.meta', ['snippet' => 'Docs'])
                ->where('serp.results.1.id', 'organic_results-0')
                ->where('serp.results.1.faviconUrl', null)
                ->where('serp.results.1.imageUrl', null)
            );
    }

    public function test_destroy(): void
    {
        $user = User::factory()->verified()->create();
        $serp = Serp::factory()->createQuietly();

        $this->actingAs($user)
            ->delete(route('serps.destroy', $serp))
            ->assertRedirect(route('serps.index'));

        $this->assertSoftDeleted($serp);
    }
}
