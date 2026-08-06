<?php

namespace App\Console\Commands;

use Exception;
use Illuminate\Console\Attributes\Description;
use Illuminate\Console\Attributes\Signature;
use Illuminate\Console\Command;
use Illuminate\Support\Facades\Redis;

#[Signature('redis:flush {connection}')]
#[Description('Flush a Redis database by connection.')]
class FlushRedis extends Command
{
    public function handle(): void
    {
        $name = $this->argument('connection');

        try {
            Redis::connection($name)->flushdb();
        } catch (Exception $e) {
            $this->warn("error flushing database [$name]: {$e->getMessage()}");
            return;
        }
        $this->info("flushed database [$name]");
    }
}
