<?php

namespace Database\Factories;

use App\Models\Bot;
use App\Models\Sitemap;
use Illuminate\Database\Eloquent\Factories\Factory;

/** @extends Factory<Sitemap> */
class SitemapFactory extends Factory
{
    public function definition(): array
    {
        $domain = $this->faker->domainName;
        return [
            'bot_id' => Bot::factory()->sitemap(),
            'domain' => $domain,
            'urls' => [
                "https://$domain",
                "https://$domain/blog",
                "https://$domain/blog/laravel",
                "https://$domain/blog/docs/getting-started",
            ],
        ];
    }

    public function deleted(): static
    {
        return $this->state(fn (array $attributes) => [
            'deleted_at' => now(),
        ]);
    }
}
