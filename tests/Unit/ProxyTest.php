<?php

namespace Tests\Unit;

use App\Models\Proxy;
use Tests\TestCase;

class ProxyTest extends TestCase
{
    public function test_to_string(): void
    {
        $proxy = new Proxy(['scheme' => 'http', 'username' => 'u', 'pass' => 'p', 'host' => 'proxy.example.com', 'port' => 8080]);

        $this->assertSame('http://u:p@proxy.example.com:8080', (string) $proxy);
    }
}
