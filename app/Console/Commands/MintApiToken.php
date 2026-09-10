<?php

namespace App\Console\Commands;

use App\Models\User;
use Illuminate\Console\Attributes\Description;
use Illuminate\Console\Attributes\Signature;
use Illuminate\Console\Command;
use Illuminate\Database\Eloquent\ModelNotFoundException;

#[Signature('api:token {email} {--name=worker}')]
#[Description('Mint a Sanctum API token for the scraper worker.')]
class MintApiToken extends Command
{
    public function handle(): void
    {
        $email = $this->argument('email');

        try {
            $user = User::query()->where('email', $email)->firstOrFail();
        } catch (ModelNotFoundException) {
            $this->warn("User not found [$email]");
            return;
        }

        $name = strval($this->option('name'));

        $token = $user->createToken($name, ['worker']);

        $this->info("token [$name] minted for [$email], copy it now - it is not shown again:");
        $this->line($token->plainTextToken);
    }
}
