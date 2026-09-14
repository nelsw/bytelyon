<?php

namespace App\Jobs;

use App\Facades\Lambda;
use DateTime;
use Illuminate\Support\Facades\Log;

class SearchBotJob extends BaseBotJob
{
    public function retryUntil(): DateTime
    {
        return now()->plus(minutes: 3);
    }

    public function handle(): void {

        $attributes = Lambda::serp(
            query: $this->bot->query,
            proxy: $this->bot->user->proxies->first(),
        );

        $this->bot->serp->update($attributes);
        $this->bot->update(['last_run_at' => now()->utc()]);

        Log::info('BotJob', [
            'id' => $this->bot->id,
            'type' => $this->bot->type->value,
            'query' => $this->bot->query,
            'result' => $attributes,
        ]);
    }
}
