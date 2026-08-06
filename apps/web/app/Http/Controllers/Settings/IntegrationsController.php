<?php

namespace App\Http\Controllers\Settings;

use Anthropic\Core\Exceptions\APIException;
use App\Http\Controllers\Controller;
use App\Http\Requests\Settings\AnthropicUpdateRequest;
use App\Http\Requests\Settings\ShopifyUpdateRequest;
use App\Services\AnthropicService;
use Illuminate\Http\RedirectResponse;
use Illuminate\Http\Request;
use Inertia\Inertia;
use Inertia\Response;

class IntegrationsController extends Controller
{
    public function edit(Request $request, AnthropicService $anthropicService): Response
    {
        $anthropic = $request->user()->anthropic;

        return Inertia::render('settings/Integrations', [
            'anthropic' => $anthropic,
            'anthropicModels' => $this->anthropicModels($anthropicService, $anthropic?->api_key),
            'shopify' => $request->user()->shopify,
        ]);
    }

    /** @return array<int, string> */
    private function anthropicModels(AnthropicService $anthropicService, ?string $apiKey): array
    {
        if (blank($apiKey)) {
            return [];
        }

        try {
            return $anthropicService->models($apiKey);
        } catch (APIException $exception) {
            report($exception);

            return [];
        }
    }

    public function updateAnthropic(AnthropicUpdateRequest $request): RedirectResponse
    {
        $request->user()->anthropic()->updateOrCreate([], $request->validated());

        Inertia::flash('toast', ['type' => 'success', 'message' => __('Anthropic settings updated.')]);

        return to_route('integrations.edit');
    }

    public function updateShopify(ShopifyUpdateRequest $request): RedirectResponse
    {
        $request->user()->shopify()->updateOrCreate([], $request->validated());

        Inertia::flash('toast', ['type' => 'success', 'message' => __('Shopify settings updated.')]);

        return to_route('integrations.edit');
    }
}
