<?php

namespace App\Console\Commands;

use App\Facades\Go;
use Illuminate\Console\Attributes\Description;
use Illuminate\Console\Attributes\Signature;
use Illuminate\Console\Command;

#[Signature('go:pub {id}')]
#[Description('Send a bot the go engine.')]
class GoPub extends Command
{
    public function handle(): void
    {
        $this->info("Sending bot ({$this->argument('id')}) to the go engine.");
        Go::pub($this->argument('id'));
        $this->info('Sent.');
    }
}
