<?php

namespace Tests\Feature\Services;

use App\Services\RssService;
use Tests\TestCase;

class RssServiceTest extends TestCase
{
    public function test_news(): void
    {
        $this->assertNotEmpty(new RssService()->news('btc forecast'));
    }
}
