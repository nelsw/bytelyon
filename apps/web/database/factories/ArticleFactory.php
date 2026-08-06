<?php

namespace Database\Factories;

use App\Models\Article;
use App\Models\Bot;
use Illuminate\Database\Eloquent\Factories\Factory;

/**
 * @extends Factory<Article>
 */
class ArticleFactory extends Factory
{
    /**
     * @return array<string, mixed>
     */
    public function definition(): array
    {
        return [
            'bot_id' => Bot::factory()->neverRun(),
            'url' => fake()->unique()->url(),
            'title' => fake()->sentence(),
            'img_alt' => fake()->words(3, true),
            'img_url' => fake()->imageUrl(),
            'description' => fake()->paragraph(),
            'body' => fake()->paragraphs(3, true),
            'source' => fake()->domainName(),
            'keywords' => [
                fake()->word(),
                fake()->word(),
                fake()->word(),
                fake()->word(),
                fake()->word(),
            ],
            'published_at' => fake()->dateTimeBetween('-1 year'),
            'publisher' => fake()->company(),
        ];
    }
}
