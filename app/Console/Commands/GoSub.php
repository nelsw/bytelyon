<?php

namespace App\Console\Commands;

use App\Facades\Go;
use Illuminate\Console\Attributes\Description;
use Illuminate\Console\Attributes\Signature;
use Illuminate\Console\Command;

#[Signature('go:sub')]
#[Description('Listen for go engine for events.')]
class GoSub extends Command
{
    public function handle(): void
    {
        $this->info('Listening to the go engine for events...');
        Go::sub(fn (string $message) => $this->info("Received:\n$message"));
        $this->info('Done listening.');
    }
}
