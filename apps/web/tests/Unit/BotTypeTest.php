<?php

namespace Tests\Unit;

use App\Enums\BotType;
use Tests\TestCase;

class BotTypeTest extends TestCase
{
    public function test_values(): void
    {
        $this->assertEquals(['news', 'search', 'sitemap'], BotType::values());
    }
}
