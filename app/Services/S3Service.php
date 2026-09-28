<?php

namespace App\Services;

use App;
use Closure;
use DateInterval;
use DateTimeInterface;
use Illuminate\Filesystem\AwsS3V3Adapter;
use Illuminate\Support\Facades\Log;
use Illuminate\Support\Facades\Storage;

class S3Service
{
    protected  AwsS3V3Adapter $client;

    public function __construct()
    {
        $this->client = Storage::disk('s3');
    }

    public function del(?string $key): bool {
        if (App::runningUnitTests()) {
            return false;
        }
        if ($key === null) {
            Log::warning("S3Service::del() called with null key");
            return false;
        }
        return $this->client->delete($key);
    }

    public function url(?string $key, Closure|DateTimeInterface|DateInterval|int $ttl = 900): ?string {
        if ($key === null) {
            Log::warning("S3Service::url() called with null key");
            return null;
        }
        return cache()->remember(
            "s3:url:$key",
            $ttl,
            fn() => $this->client->temporaryUrl($key, now()->addSeconds($ttl))
        );
    }
}
