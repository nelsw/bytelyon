<?php

namespace Database\Factories;

use App\Models\Proxy;
use App\Models\User;
use Illuminate\Database\Eloquent\Factories\Factory;
use Illuminate\Support\Carbon;

class ProxyFactory extends Factory
{
    protected $model = Proxy::class;

    public function definition(): array
    {
        return [
            'server' => $this->faker->word(),
            'username' => $this->faker->userName(),
            'password' => bcrypt($this->faker->password()),
            'bypass' => $this->faker->word(),
            'name' => $this->faker->name(),
            'created_at' => Carbon::now(),
            'updated_at' => Carbon::now(),
        ];
    }
}
