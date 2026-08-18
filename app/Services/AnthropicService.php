<?php

namespace App\Services;

use Anthropic\Client;
use Anthropic\Core\Exceptions\APIException;
use Anthropic\Messages\Message;
use Anthropic\Messages\Model;
use Anthropic\Models\ModelInfo;
use Illuminate\Container\Attributes\Singleton;
use Illuminate\Support\Arr;
use Illuminate\Support\Facades\Log;
use Illuminate\Support\Str;

#[Singleton]
readonly class AnthropicService
{
    private function client(string $apiKey): Client
    {
        return new Client(apiKey: $apiKey);
    }

    /** @throws APIException */
    public function models(string $apiKey): array
    {
        /** @var ModelInfo[]|array|null $data */
        $data = $this->client($apiKey)->models->list()->data;

        if ($data === null) {
            return [];
        }

        return Arr::map($data, fn (ModelInfo $info): string => $info->id);
    }

    /** @throws APIException */
    public function prompt(
        string $apiKey,
        array $messages,
        int $maxTokens = 1024,
        Model|string $model = Model::CLAUDE_FABLE_5,
        array|string|null $system = null,
        bool $html = false,
    ): string {

        /** @var Message $message */
        $message = $this->client($apiKey)->messages->create(
            maxTokens: $maxTokens,
            messages: $messages,
            model: $model,
            system: $system,
        );

        $text = collect($message->content)
            ->filter(fn ($item) => $item['type'] === 'text')
            ->pluck('text')
            ->implode(' ');

        if ($html) {
            $text = Str::markdown($text);
            $text = str_replace("\n", '', $text);
        }

        Log::debug('Anthropic::prompt', [
            'messages' => $messages,
            'system' => $system,
            'text' => $text,
        ]);

        return $text;
    }
}
