<?php

namespace App\Actions;

use App\Concerns\ArticleValidationRules;

class UpdateArticle
{
    use ArticleValidationRules;

    public function __invoke(): void {}
}
