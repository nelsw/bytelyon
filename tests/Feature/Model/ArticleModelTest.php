<?php

namespace Tests\Feature\Model;

use App\Models\Article;
use Illuminate\Support\Facades\App;
use Tests\TestCase;

class ArticleModelTest extends TestCase
{
    public function test_create(): void
    {
        $articles = Article::factory()
            ->count(3)
            ->create();

        if (App::isLocal()) {
            dump($articles->toPrettyJson());
        }

        $this->assertDatabaseHas($articles);
    }
}
