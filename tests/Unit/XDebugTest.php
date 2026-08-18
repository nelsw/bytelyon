<?php

namespace Tests\Unit;

use Tests\TestCase;

class XDebugTest extends TestCase
{
    public function test_breakpoints(): void
    {
        $expected = fake()->boolean();
        if ($expected) {
            $actual = true;
        } else {
            $actual = false;
        }
        $this->assertEquals($expected, $actual);
    }
}
