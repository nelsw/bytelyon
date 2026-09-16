<?php

namespace Tests\Unit\Traits;

use App\Support\Sqs\Page;
use Tests\TestCase;

class HasLinksTest extends TestCase
{
    public function test_has_links()
    {
        $contentKey = 'worker-scrapes/sitemap/4066bff0-dbbf-4d22-b062-69d002298598/www_ubicquia_com/index-fbca31bf74.html';
        $page = new Page('https://www.ubicquia.com', $contentKey);
        $links = $page->links();
        $this->assertGreaterThan(0, count($links));
    }
}
