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

    /**
     * Enqueues a scrape job onto the shared bytelyon-scrape-jobs SQS queue
     * instead of invoking a Lambda function synchronously. One or more
     * local-network workers (docker/worker/worker.py) long-poll that queue
     * and run the actual CloakBrowser scrape themselves — Google trusts
     * residential/home-network egress far more than any datacenter IP, AWS
     * Lambda's own included (see docker/worker/handlers/serp.py's module
     * docstring for the full rationale). This makes the job async:
     * `Serp::data` etc. are populated later, whenever a worker reports back
     * via POST /api/scrape-jobs/serp/{serp}/complete, not by the time this
     * method returns.
     *
     * No per-user Proxy is needed for this path — the worker's own
     * DataImpulse proxy credentials are baked into its Docker image, not
     * looked up per-user.
     */
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
