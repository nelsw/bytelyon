<?php

namespace Database\Factories;

use App\Enums\BotType;
use App\Enums\FrequencyType;
use App\Models\Bot;
use App\Models\User;
use Illuminate\Database\Eloquent\Factories\Factory;

/**
 * @extends Factory<Bot>
 */
class BotFactory extends Factory
{
    /**
     * @return array<string, mixed>
     */
    public function definition(): array
    {
        $type = fake()->randomElement(BotType::values());
        if (BotType::from($type) === BotType::Sitemap) {
            $query = fake()->domainName();
        } else {
            $query = fake()->sentence();
        }
        return [
            'user_id' => User::factory(),
            'blacklist' => "foo\nbar\nbaz",
            'headless' => fake()->boolean(),
            'frequency' => fake()->randomElement(FrequencyType::values()),
            'query' => $query,
            'type' => $type,
            'last_run_at' => now()->subHours(fake()->randomDigitNotNull()),
            'enabled' => fake()->boolean(),
        ];
    }

    public function enabled(): static
    {
        return $this->state(fn (array $attributes) => [
            ...$attributes,
            ...['enabled' => true],
        ]);
    }

    public function neverRun(): static
    {
        return $this->state(fn (array $attributes) => [
            ...$attributes,
            ...['last_run_at' => null],
        ]);
    }

    public function news(): static
    {
        return $this->state(fn (array $attributes) => [
            ...$attributes,
            ...[
                'type' => BotType::News,
                'query' => fake()->sentence(),
            ],
        ]);
    }

    public function search(): static
    {
        return $this->state(fn (array $attributes) => [
            ...$attributes,
            ...[
                'type' => BotType::Search,
                'query' => fake()->sentence(),
            ],
        ]);
    }

    public function sitemap(): static
    {
        return $this->state(fn (array $attributes) => [
            ...$attributes,
            ...[
                'type' => BotType::Sitemap,
                'query' => fake()->domainName(),
            ],
        ]);
    }
}
