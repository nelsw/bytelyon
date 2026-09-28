<?php

namespace Tests\Feature\Services;

use App\Services\S3Service;
use Illuminate\Filesystem\AwsS3V3Adapter;
use Illuminate\Support\Facades\Cache;
use Illuminate\Support\Facades\Storage;
use Mockery;
use Mockery\MockInterface;
use Tests\TestCase;

class S3ServiceTest extends TestCase
{
    private MockInterface $disk;

    protected function setUp(): void
    {
        parent::setUp();

        $this->disk = Mockery::mock(AwsS3V3Adapter::class);
        Storage::shouldReceive('disk')->with('s3')->andReturn($this->disk);
    }

    public function test_del_is_a_no_op_while_testing(): void
    {
        $this->disk->shouldNotReceive('delete');

        $this->assertFalse(new S3Service()->del('some/key.png'));
    }

    public function test_url_returns_null_for_a_null_key(): void
    {
        $this->assertNull(new S3Service()->url(null));
    }

    public function test_url_is_generated_once_and_cached(): void
    {
        $key = 'screens/'.fake()->uuid().'.png';
        Cache::forget("s3:url:$key");

        $this->disk->shouldReceive('temporaryUrl')->once()->andReturn('https://signed.example.com/x');

        $service = new S3Service;
        $this->assertSame('https://signed.example.com/x', $service->url($key));
        $this->assertSame('https://signed.example.com/x', $service->url($key));
    }
}
