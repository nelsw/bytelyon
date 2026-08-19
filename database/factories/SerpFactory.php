<?php

namespace Database\Factories;

use App\Models\Bot;
use App\Models\Serp;
use Illuminate\Database\Eloquent\Factories\Factory;

/**
 * @extends Factory<Serp>
 */
class SerpFactory extends Factory
{
    /**
     * @return array<string, mixed>
     */
    public function definition(): array
    {
        return [
            'bot_id' => Bot::factory()->search(),
            'query' => $this->faker->sentence,
            'screenshot_key' => $this->faker->uuid.'.png',
            'content_key' => $this->faker->uuid.'.html',
            'data' => [
                'results' => [
                    ['title' => $this->faker->sentence, 'link' => $this->faker->url],
                    ['title' => $this->faker->sentence, 'link' => $this->faker->url],
                ],
            ],
        ];
    }

    public function deleted(): static
    {
        return $this->state(fn (array $attributes) => [
            'deleted_at' => now(),
        ]);
    }

    public function configure(): static
    {
        return $this->afterCreating(function (Serp $serp) {
            Serp::withTrashed()
                ->where('bot_id', $serp->bot_id)
                ->whereKeyNot($serp->id)
                ->forceDelete();
        });
    }
}
