<?php

namespace Database\Factories;

use App\Models\Page;
use App\Models\Serp;
use App\Models\Sitemap;
use Illuminate\Database\Eloquent\Factories\Factory;

/** @extends Factory<Page> */
class PageFactory extends Factory
{
    public function definition(): array
    {
        $domain = $this->faker->domainName;
        return [
            'domain' => $domain,
            'url' => "https://$domain",
            'title' => $this->faker->sentence,
            'screenshot_key' => $this->faker->uuid,
            'meta' => [
                $this->faker->word => [
                    $this->faker->word,
                    $this->faker->word,
                ],
                $this->faker->word => [
                    $this->faker->word,
                    $this->faker->word,
                ],
            ],
        ];
    }

    public function withRelation(string $class, int $id): static
    {
        return $this->state([
            'pageable_id' => $id,
            'pageable_type' => $class,
        ]);
    }

    public function serp(): static
    {
        return $this->withRelation(Serp::class, Serp::factory()->create()->id);
    }

    public function sitemap(): static
    {
        return $this->withRelation(Sitemap::class, Sitemap::factory()->create()->id);
    }
}
