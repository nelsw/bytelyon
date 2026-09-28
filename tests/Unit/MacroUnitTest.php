<?php

namespace Tests\Unit;

use App\Models\User;
use Illuminate\Support\Arr;
use Illuminate\Support\Facades\URL;
use Illuminate\Support\Fluent;
use Tests\TestCase;

class MacroUnitTest extends TestCase
{
    public function test_arr_to_camel_case(): void
    {
        $this->assertSame([], Arr::toCamelCase(null));
        $this->assertSame(
            ['fooBar' => 1, 'nestedKey' => ['innerKey' => 2]],
            Arr::toCamelCase(['foo_bar' => 1, 'nested_key' => ['inner_key' => 2]]),
        );
        $this->assertSame(
            ['name' => 'Jane', 'imgUrl' => null],
            Arr::toCamelCase(new User(['name' => 'Jane', 'img_url' => null])),
        );
    }

    public function test_url_clean(): void
    {
        $this->assertSame('https://laravel.com/docs', URL::clean('  https://laravel.com/docs/ '));
        $this->assertSame('', URL::clean(null));
    }

    public function test_url_domain(): void
    {
        $this->assertSame('laravel.com', URL::domain('https://laravel.com/docs/13.x'));
        $this->assertSame('laravel.com', URL::domain('https://www.laravel.com'));
        $this->assertSame('example.com', URL::domain('example.com'));
        $this->assertSame('', URL::domain(null));
    }

    public function test_url_to_favicon(): void
    {
        $this->assertSame(
            'https://www.google.com/s2/favicons?domain=bytelyon.com&sz=64',
            URL::toFavicon(null),
        );
        $this->assertSame(
            'https://www.google.com/s2/favicons?domain=laravel.com&sz=32',
            URL::toFavicon('https://www.laravel.com/docs', 32),
        );
    }

    public function test_fluent_of(): void
    {
        $fluent = Fluent::of('{"a":1,"b":{"c":2}}');
        $this->assertSame(1, $fluent->get('a'));
        $this->assertSame(2, $fluent->get('b.c'));
    }
}
