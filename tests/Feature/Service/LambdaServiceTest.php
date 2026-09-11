<?php

namespace Tests\Feature\Service;

use App\Models\Proxy;
use App\Services\LambdaService;
use Illuminate\Support\Facades\App;
use Illuminate\Support\Facades\Log;
use Tests\TestCase;

class LambdaServiceTest extends TestCase
{
    private LambdaService $service;

    protected function setUp(): void
    {
        parent::setUp();
        $this->service = new LambdaService;
    }

    public function test_news(): void
    {
        $out = $this->service->news([
            'https://www.theguardian.com/world/2026/sep/09/iran-claims-to-have-attacked-10-ships-near-strait-of-hormuz-after-us-strikes',
            'https://www.yahoo.com/news/politics/articles/trump-promises-end-iran-war-100000311.html\?guccounter\=1\&guce_referrer\=aHR0cHM6Ly9mb3JnZS5sYXJhdmVsLmNvbS8\&guce_referrer_sig\=AQAAAEE1mljUCq4KU0tkmHtzbVFCkjR-5vtoByRGt6cUT-WluSOr_wzwhLfcgOYHUM8Q7dRjFZvswqtv9HTQz8lPD8OKyoTjLySlN_iHcm35WeXTD0k8HPmW2zfx0F8-0cRydlc_MiUGDzPcebUh8usWuvDRbLbZ44jwcniHiWOfk0-Q',
        ]);
        $this->assertIsArray($out);
        Log::debug('LambdaServiceTest#test_news', $out);
    }

    public function test_search(): void
    {
        $out = $this->service->serp('sailing blocks', new Proxy([
            'server' => 'socks5://gw.dataimpulse.com:824',
            'bypass' => '.google.com',
            'username' => '2b9bc9b578274c5e9c8f__cr.us',
            'password' => 'a1d0893a3f201adf',
        ]));
        $this->assertIsArray($out);
        if (! App::isProduction()) {
            dump($out);
        }
    }

    public function test_page_scraper(): void
    {
        $out = $this->service->page('https://li-fire.com');
        $this->assertIsArray($out);
        if (! App::isProduction()) {
            dump($out);
        }
    }
}
