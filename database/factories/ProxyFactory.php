<?php

namespace Database\Factories;

use App\Models\Proxy;
use App\Models\User;
use Illuminate\Database\Eloquent\Factories\Factory;

/** @extends Factory<Proxy> */
class ProxyFactory extends Factory
{
    protected $model = Proxy::class;

    public function definition(): array
    {
        return [
            'scheme' => $this->faker->randomElement(['http', 'https', 'socks5']),
            'host' => $this->faker->word(),
            'port' => $this->faker->numberBetween(1024, 9000),
            'username' => $this->faker->userName(),
            'pass' => bcrypt($this->faker->password()),
            'bypass' => $this->faker->word(),
            'name' => $this->faker->name(),
            'created_at' => now(),
            'updated_at' => now(),
            'user_id' => User::factory()->verified(),
        ];
    }

    public function default(): static
    {
        return $this->state(fn (array $attributes) => [
            ...$attributes,
            ...config('services.proxy'),
        ]);
    }
}
