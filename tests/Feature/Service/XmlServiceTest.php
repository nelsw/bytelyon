<?php

namespace Tests\Feature\Service;

use App\Services\XmlService;
use Tests\TestCase;

class XmlServiceTest extends TestCase
{
    private XmlService $service;

    protected function setUp(): void
    {
        parent::setUp();
        $this->service = new XmlService();
    }

    public function testFetch():void
    {
        $url = "https://www.bing.com/news/search";
        $query = [
            'q' => 'btc forecast',
            'format' => 'rss',
        ];

        $xml = $this->service->fetch($url, $query);
        $this->assertIsNotBool($xml);
    }
}
