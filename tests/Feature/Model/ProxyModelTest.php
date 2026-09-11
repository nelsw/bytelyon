<?php

namespace Tests\Feature\Model;

use App\Models\Proxy;
use App\Models\User;
use Tests\TestCase;

class ProxyModelTest extends TestCase
{
    public function test_proxy_user_relation(): void
    {
        $user = User::factory()->create();

        $proxy1 = Proxy::factory()->makeOne();
        $proxy2 = Proxy::factory()->createOne(['user_id' => $user->id]);

        $user->proxies()->create($proxy1->toArray());

        $this->assertDatabaseHas('proxies', $proxy1->attributesToArray());
        $this->assertDatabaseHas('proxies', $proxy2->attributesToArray());
        $this->assertCount(2, $user->proxies);
    }
}
