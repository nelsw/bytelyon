<?php

namespace App\Actions\Model;

use App\Concerns\ArticleValidationRules;
use App\Models\Article;
use Illuminate\Support\Facades\Log;

class UpdateArticle
{
    use ArticleValidationRules;

    /** @param array<string> $keywords */
    public function __invoke(
        Article $article,
        string $body,
        string $description,
        string $imgAlt,
        string $imgUrl,
        array $keywords,
    ): void {

        when(empty($article->description), fn () => $article->setAttribute('description', $description));
        when(empty($article->img_alt), fn () => $article->setAttribute('img_alt', $imgAlt));
        when(empty($article->img_url), fn () => $article->setAttribute('img_url', $imgUrl));
        $article->setAttribute('body', $body);
        $article->setAttribute('keywords', $keywords);
        $article->save();

        Log::debug('Article updated', $article->toArray());
    }
}
