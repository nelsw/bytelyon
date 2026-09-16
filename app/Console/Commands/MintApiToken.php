<?php

namespace App\Console\Commands;

use App\Models\User;
use Illuminate\Console\Attributes\Description;
use Illuminate\Console\Attributes\Signature;
use Illuminate\Console\Command;
use Illuminate\Database\Eloquent\ModelNotFoundException;
use Throwable;

#[Signature('api:token {email} {--name=worker}')]
#[Description('Mint a Sanctum API token for the scraper worker.')]
class MintApiToken extends Command
{
    /** @throws Throwable */
    public function handle(): void
    {
        $email = $this->argument('email');

        try {
            $user = User::query()->where('email', $email)->firstOrFail();
        } catch (ModelNotFoundException) {
            $this->fail("User not found where email=[$email]");
        }

        $name = strval($this->option('name'));

        $this->info("token [$name] minted for [$email], copy it now - it is not shown again:");
        $this->line($user->createToken($name, ['worker'])->plainTextToken);
    }
}
