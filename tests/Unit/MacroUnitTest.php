<?php

namespace Tests\Unit;

use Illuminate\Support\Facades\URL;
use Tests\TestCase;

class MacroUnitTest extends TestCase
{
    public function test_ur_l_to_domain(): void
    {
        $url = 'https://laravel.com/docs/13.x';
        $expected = 'laravel.com';
        $actual = URL::toDomain($url);
        $this->assertEquals($expected, $actual);
    }
}
