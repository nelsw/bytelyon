<?php

namespace Database\Seeders;

use App\Models\Bot;
use App\Models\User;
use Illuminate\Database\Console\Seeds\WithoutModelEvents;
use Illuminate\Database\Seeder;

class DatabaseSeeder extends Seeder
{
    use WithoutModelEvents;

    public function run(): void
    {
        $user = User::factory()
            ->verified()
            ->create([
                'name' => 'Test User',
                'email' => 'test@example.com',
            ]);

        Bot::factory()
            ->for($user)
            ->create();
    }
}
