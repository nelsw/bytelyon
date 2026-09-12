<?php

namespace Database\Factories;

use App\Models\Proxy;
use Illuminate\Database\Eloquent\Factories\Factory;
use Illuminate\Support\Carbon;

class ProxyFactory extends Factory
{
    protected $model = Proxy::class;

    public function definition(): array
    {
        return [
            'scheme' => $this->faker->randomElement(['http', 'https', 'socks5']),
            'host' => $this->faker->word(),
            'port' => $this->faker->numberBetween(1024, 9000),
            'user' => $this->faker->userName(),
            'pass' => bcrypt($this->faker->password()),
            'bypass' => $this->faker->word(),
            'name' => $this->faker->name(),
            'created_at' => Carbon::now(),
            'updated_at' => Carbon::now(),
        ];
    }
}
