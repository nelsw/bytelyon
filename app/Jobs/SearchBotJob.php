<?php

namespace App\Jobs;

use App\Facades\Sqs;
use DateTime;
use Illuminate\Support\Facades\Log;

class SearchBotJob extends BaseBotJob
{
    public function retryUntil(): DateTime
    {
        return now()->plus(minutes: 3);
    }

    public function handle(): void
    {
        Sqs::enqueueScrape('serp', $this->bot->serp->id, ['query' => $this->bot->query]);

        $this->bot->update(['last_run_at' => now()->utc()]);

        Log::info('BotJob: enqueued', [
            'id' => $this->bot->id,
            'type' => $this->bot->type->value,
            'query' => $this->bot->query,
            'serp_id' => $this->bot->serp->id,
        ]);
    }
}
