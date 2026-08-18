<?php

namespace App\Http\Controllers;

use Anthropic\Core\Exceptions\APIException;
use Anthropic\Messages\Model as AnthropicModel;
use App\Concerns\ArticleValidationRules;
use App\Exceptions\ShopifyException;
use App\Models\Article;
use App\Models\Bot;
use App\Services\AnthropicService;
use App\Services\ShopifyService;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\RedirectResponse;
use Illuminate\Http\Request;
use Illuminate\Routing\Attributes\Controllers\Authorize;
use Illuminate\Support\Facades\Log;
use Illuminate\Support\Str;
use Inertia\Inertia;
use Inertia\Response;

class ArticleController extends Controller
{
    use ArticleValidationRules;

    #[Authorize('view', 'bot')]
    public function index(Bot $bot): Response
    {
        return Inertia::render('articles/Index', [
            'bot' => $bot,
            'articles' => $bot->articles()
                ->latest('published_at')
                ->get()
                ->map(Article::row()),
        ]);
    }

    #[Authorize('view', 'bot')]
    public function show(Bot $bot, Article $article): Response
    {
        return Inertia::render('articles/Show', [
            'bot' => $bot,
            'article' => $article,
        ]);
    }

    #[Authorize('update', 'bot')]
    public function edit(Request $request, Bot $bot, Article $article): Response
    {
        return Inertia::render('articles/Edit', [
            'bot' => $bot,
            'article' => $article,
            'botOptions' => $request->user()->bots()
                ->orderBy('query')
                ->get()
                ->map(fn (Bot $b) => [
                    'value' => (string) $b->getKey(),
                    'label' => $b->query,
                ])
                ->all(),
        ]);
    }

    #[Authorize('update', 'bot')]
    public function update(Request $request, Bot $bot, Article $article): RedirectResponse
    {
        $article->update($request->validate($this->articleRules()));

        Inertia::flash('toast', ['type' => 'success', 'message' => __('Article updated.')]);

        return to_route('articles.edit', ['bot' => $bot, 'article' => $article]);
    }

    #[Authorize('update', 'bot')]
    public function assist(Request $request, Bot $bot, Article $article, AnthropicService $anthropicService): JsonResponse
    {
        $validated = $request->validate([
            'prompt' => ['required', 'string', 'max:4000'],
            'system' => ['nullable', 'string', 'max:2000'],
            'body' => ['nullable', 'string'],
        ]);

        Log::debug('ArticleController::assist', $validated);

        $anthropic = $request->user()->anthropic;

        if (blank($anthropic?->api_key)) {
            return response()->json([
                'message' => __('Add an Anthropic API key in integration settings before using AI assist.'),
            ], 422);
        }

        $system = $validated['system'];
        if ($system !== null) {
            $system = trim($system);
        }

        try {
            $text = $anthropicService->prompt(
                apiKey: $anthropic->api_key,
                messages: [[
                    'role' => 'user',
                    'content' => trim(
                        "Instruction: {$validated['prompt']}\n\n".
                        "Current article body (HTML):\n{$validated['body']}"
                    ),
                ]],
                model: $anthropic->default_model ?: AnthropicModel::CLAUDE_FABLE_5,
                system: $system,
            );
        } catch (APIException $exception) {
            report($exception);

            return response()->json([
                'message' => __('The AI assist request failed. Please try again.'),
            ], 502);
        }

        return response()->json([
            'html' => Str::markdown($text),
        ]);
    }

    #[Authorize('update', 'bot')]
    public function publish(Request $request, Bot $bot, Article $article, ShopifyService $shopifyService): RedirectResponse
    {
        $shopify = $request->user()->shopify;

        if (blank($shopify)) {
            Inertia::flash('toast', ['type' => 'error', 'message' => __('Add your Shopify store details in integration settings before publishing.')]);

            return to_route('articles.edit', ['bot' => $bot, 'article' => $article]);
        }

        try {
            $shopifyService->createArticle(
                shopify: $shopify,
                body: $article->body ?? '',
                title: $article->title,
                tags: $article->keywords ?? [],
                publishedAt: $article->published_at,
                imageAlt: $article->img_alt,
                imageUrl: $article->img_url,
                summary: $article->description,
            );
        } catch (ShopifyException $exception) {
            report($exception);

            Inertia::flash('toast', ['type' => 'error', 'message' => __('Publishing to Shopify failed. Please try again.')]);

            return to_route('articles.edit', ['bot' => $bot, 'article' => $article]);
        }

        Inertia::flash('toast', ['type' => 'success', 'message' => __('Article published to Shopify.')]);

        return to_route('articles.edit', ['bot' => $bot, 'article' => $article]);
    }

    #[Authorize('update', 'bot')]
    public function destroy(Bot $bot, Article $article): RedirectResponse
    {
        $article->delete();

        Inertia::flash('toast', ['type' => 'success', 'message' => __('Article deleted.')]);

        return to_route('articles.index', $bot);
    }
}
