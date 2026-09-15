"""Generic SQS polling worker for all bytelyon scrape job types (serp, news,
sitemap, ...) — the single consolidated location this whole worker fleet
lives in, replacing what used to be a per-job-type Lambda function.

Why local workers instead of Lambda at all: Google (and to a lesser extent
other sites) trusts residential/home-network egress far more than any
datacenter IP, Lambda's own included. Running the actual browser on a home
network, driven by jobs pulled from a shared SQS queue, sidesteps that
entirely — and any number of machines can run this same worker against the
same queue for free, horizontal, pull-based scaling (SQS's visibility
timeout guarantees only one worker processes a given message at a time).

Every handler under handlers/ returns the exact same shape:
    {"url": ..., "screenshot_key": ..., "content_key": ...}
No parsing, no per-job business logic lives in Python at all — see
handlers/serp.py and handlers/generic.py's own docstrings. All of that
(SERP structure, article body extraction, sitemap link discovery) lives in
the main Laravel app (app/Support/Html.php, app/Data/Html/Page.php,
app/Support/Serp.php), reached via one callback endpoint per job.

Flow, per message:
    1. Long-poll receive from SQS (`SCRAPE_JOBS_QUEUE_URL`).
    2. Parse `{"type": "serp"|"news"|"sitemap", "id": int, ...}` from the
       message body. `type` picks the handler; `id` identifies which
       record to update on the Laravel side (its meaning is type-specific
       — see routes/api.php and ScrapeJobController for what `id` means
       per type). Any other fields (e.g. sitemap's `depth`) are opaque to
       this worker — it doesn't need to understand them, just carry them
       through to the callback unchanged.
    3. Run the matching handler's `run(job)` -> {url, screenshot_key,
       content_key}.
    4. POST {**passthrough fields, url, screenshot_key, content_key} to
       POST {LARAVEL_BASE_URL}/api/scrape-jobs/{type}/{id}/complete,
       Authorization: Bearer {LARAVEL_WORKER_TOKEN} (a Sanctum personal
       access token with the `worker` ability, issued from
       Settings > API Tokens).
    5. On success, delete the SQS message. On any failure (scrape or
       callback), leave it alone — SQS's visibility timeout expires and
       redelivers it (up to the queue's maxReceiveCount) before it lands in
       the DLQ, so a transient failure gets retried automatically without
       any retry logic needed here.

Required environment variables:
    AWS_ACCESS_KEY_ID / AWS_SECRET_ACCESS_KEY   scoped IAM credentials
                                                 (sqs:ReceiveMessage/DeleteMessage
                                                 on the queue below,
                                                 s3:PutObject on
                                                 bytelyon-private/worker-scrapes/*)
    AWS_DEFAULT_REGION                          default "us-east-1"
    SCRAPE_JOBS_QUEUE_URL                        required, full SQS queue URL
    LARAVEL_BASE_URL                            default "http://host.docker.internal"
                                                 (reaches the Sail app's port
                                                 80 on the Docker host from
                                                 inside this container, on
                                                 Docker Desktop for Mac —
                                                 override for other setups)
    LARAVEL_WORKER_TOKEN                        required, Sanctum PAT with
                                                 the `worker` ability
"""

from __future__ import annotations

import json
import logging
import os
import sys
import time

import boto3
import requests
from handlers import generic, serp

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s %(levelname)s %(message)s",
    stream=sys.stdout,
)
logger = logging.getLogger("worker")

QUEUE_URL = os.environ["SCRAPE_JOBS_QUEUE_URL"]
LARAVEL_BASE_URL = os.environ.get(
    "LARAVEL_BASE_URL", "http://host.docker.internal"
).rstrip("/")
LARAVEL_WORKER_TOKEN = os.environ["LARAVEL_WORKER_TOKEN"]

sqs = boto3.client("sqs", region_name=os.environ.get("AWS_DEFAULT_REGION", "us-east-1"))

_HANDLERS = {
    "serp": serp.run,
    "news": generic.run,
    "sitemap": generic.run,
}


def _report_result(job_type: str, job_id: int, passthrough: dict, result: dict) -> None:
    url = f"{LARAVEL_BASE_URL}/api/scrape-jobs/{job_type}/{job_id}/complete"
    resp = requests.post(
        url,
        json={**passthrough, **result},
        headers={"Authorization": f"Bearer {LARAVEL_WORKER_TOKEN}"},
        timeout=15,
    )
    resp.raise_for_status()


def _handle_message(message: dict) -> None:
    receipt_handle = message["ReceiptHandle"]
    try:
        job = json.loads(message["Body"])
        job_type = str(job["type"])
        job_id = int(job["id"])
        handler = _HANDLERS[job_type]
    except Exception:
        logger.exception(
            "malformed/unknown job message, leaving for redrive: %s",
            message.get("Body"),
        )
        return

    # Everything except type/id/headless rides through to the callback
    # unchanged (e.g. sitemap's `depth`) — this worker doesn't need to
    # understand it, only the handler (for its own inputs, e.g. `query` or
    # `url`) and the Laravel callback (for continuing the workflow) do.
    passthrough = {k: v for k, v in job.items() if k not in ("type", "id", "headless")}

    logger.info("job received: type=%s id=%s", job_type, job_id)
    try:
        result = handler(job)
        logger.info(
            "job scraped: type=%s id=%s url=%s", job_type, job_id, result["url"]
        )
        _report_result(job_type, job_id, passthrough, result)
    except Exception:
        logger.exception(
            "job failed (type=%s id=%s) — leaving message for SQS redrive/DLQ",
            job_type,
            job_id,
        )
        return

    sqs.delete_message(QueueUrl=QUEUE_URL, ReceiptHandle=receipt_handle)
    logger.info("job complete: type=%s id=%s", job_type, job_id)


def main() -> None:
    logger.info("worker starting, polling %s", QUEUE_URL)
    while True:
        try:
            resp = sqs.receive_message(
                QueueUrl=QUEUE_URL,
                MaxNumberOfMessages=1,
                WaitTimeSeconds=20,
            )
        except Exception:
            logger.exception("SQS receive_message failed, backing off 5s")
            time.sleep(5)
            continue

        for message in resp.get("Messages", []):
            _handle_message(message)


if __name__ == "__main__":
    main()
