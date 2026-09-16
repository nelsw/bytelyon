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
    1. Long-poll receive from SQS (`--queue-url` / `SCRAPE_JOBS_QUEUE_URL`).
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
       POST {laravel_base_url}/api/scrape-jobs/{type}/{id}/complete,
       Authorization: Bearer {worker_token} (a Sanctum personal access
       token with the `worker` ability, issued from Settings > API Tokens).
    5. On success, delete the SQS message. On any failure (scrape or
       callback), leave it alone — SQS's visibility timeout expires and
       redelivers it (up to the queue's maxReceiveCount) before it lands in
       the DLQ, so a transient failure gets retried automatically without
       any retry logic needed here.

Configuration — every setting has both a CLI flag and an env var; the flag
wins if both are given. Env vars are the natural fit for `docker run -e`;
CLI flags are there for running this as a plain script, or for one-off
overrides without touching a container's env:

    --queue-url          SCRAPE_JOBS_QUEUE_URL   required, full SQS queue URL
    --laravel-base-url   LARAVEL_BASE_URL         default "http://host.docker.internal"
    --worker-token       LARAVEL_WORKER_TOKEN     required, Sanctum PAT with
                                                   the `worker` ability
    --aws-region         AWS_DEFAULT_REGION       default "us-east-1"

    AWS_ACCESS_KEY_ID / AWS_SECRET_ACCESS_KEY (env-only, standard boto3
    credential resolution — no CLI flag; don't put secrets on a command
    line) scoped IAM credentials (sqs:ReceiveMessage/DeleteMessage on the
    queue above, s3:PutObject on bytelyon-private/worker-scrapes/*).

Local dev vs production — use a *separate queue per environment*, not just
a different --laravel-base-url on an otherwise-shared queue. A job's `id`
only means anything against the database it was enqueued from; a worker
pointed at the wrong environment's Laravel app would either 404 or, worse,
silently complete a same-numbered but unrelated record. Point each worker
at the one queue *and* Laravel URL pair that actually go together:

    # local dev (Sail on this machine)
    python3 worker.py \\
        --queue-url https://sqs.us-east-1.amazonaws.com/<account>/bytelyon-scrape-jobs-dev \\
        --laravel-base-url http://host.docker.internal \\
        --worker-token <dev Sanctum token>

    # production
    python3 worker.py \\
        --queue-url https://sqs.us-east-1.amazonaws.com/<account>/bytelyon-scrape-jobs \\
        --laravel-base-url https://bytelyon.com \\
        --worker-token <production Sanctum token>

See INSTRUCTIONS.md for how to provision the dev queue (same shape as the
production one, just a second queue + DLQ pair).
"""

from __future__ import annotations

import argparse
import json
import logging
import os
import sys
import time
from dataclasses import dataclass

import boto3
import requests
from handlers import generic, serp


class _ColorFormatter(logging.Formatter):
    """Colors just the level name, so INFO/WARNING/ERROR/etc. are easy to
    pick out at a glance in a terminal, without touching the message itself
    (so it stays plain, greppable text). Falls back to no color at all when
    stdout isn't a TTY (e.g. piped to a file or a log collector), since ANSI
    codes would just show up as garbage escape sequences there.
    """

    _COLORS = {
        logging.DEBUG: "\033[36m",  # cyan
        logging.INFO: "\033[32m",  # green
        logging.WARNING: "\033[33m",  # yellow
        logging.ERROR: "\033[31m",  # red
        logging.CRITICAL: "\033[1;37;41m",  # bold white on red bg
    }
    _ABBREVIATIONS = {
        logging.DEBUG: "DBG",
        logging.INFO: "INF",
        logging.WARNING: "WAR",
        logging.ERROR: "ERR",
        logging.CRITICAL: "FAT",
    }
    _RESET = "\033[0m"

    def __init__(self, fmt: str, *, colorize: bool) -> None:
        super().__init__(fmt, datefmt="%H:%M")
        self._colorize = colorize

    def format(self, record: logging.LogRecord) -> str:
        levelname = self._ABBREVIATIONS.get(record.levelno, record.levelname)
        if self._colorize:
            color = self._COLORS.get(record.levelno, "")
            if color:
                levelname = f"{color}{levelname}{self._RESET}"
        if levelname != record.levelname:
            record = logging.makeLogRecord(record.__dict__)
            record.levelname = levelname
        return super().format(record)


# Not gated on sys.stdout.isatty(): this worker normally runs inside a
# plain (non-tty) Docker container, piping stdout straight to the log
# driver — isatty() would be False there even though whoever's actually
# reading the logs (`docker compose logs -f`, a terminal) is a real,
# color-capable TTY on the *other* end of that pipe. Default to color on;
# respect the standard NO_COLOR convention (https://no-color.org/) for
# anyone piping logs somewhere that can't render ANSI codes.
_handler = logging.StreamHandler(stream=sys.stdout)
_handler.setFormatter(
    _ColorFormatter(
        "%(asctime)s %(levelname)s %(message)s",
        colorize=not os.environ.get("NO_COLOR"),
    )
)
logging.basicConfig(level=logging.INFO, handlers=[_handler])
logger = logging.getLogger("worker")

_HANDLERS = {
    "serp": serp.run,
    "news": generic.run,
    "sitemap": generic.run,
}


@dataclass(frozen=True)
class Config:
    queue_url: str
    laravel_base_url: str
    worker_token: str
    aws_region: str


def _parse_config(argv: list[str] | None = None) -> Config:
    parser = argparse.ArgumentParser(
        description="bytelyon scrape-jobs SQS worker (serp/news/sitemap)"
    )
    parser.add_argument(
        "--queue-url",
        default=None,
        help="SQS queue URL to poll. Env: SCRAPE_JOBS_QUEUE_URL. Required "
        "(one or the other).",
    )
    parser.add_argument(
        "--laravel-base-url",
        default=None,
        help="Base URL of the Laravel app to POST callbacks to — e.g. "
        "https://bytelyon.com in production, http://host.docker.internal "
        "for local dev (reaches the Sail app's port 80 on the Docker host "
        "from inside this container, on Docker Desktop for Mac). Env: "
        "LARAVEL_BASE_URL. Default: http://host.docker.internal.",
    )
    parser.add_argument(
        "--worker-token",
        default=None,
        help="Sanctum personal access token with the `worker` ability "
        "(Settings > API Tokens). Env: LARAVEL_WORKER_TOKEN. Required "
        "(one or the other).",
    )
    parser.add_argument(
        "--aws-region",
        default=None,
        help="AWS region the SQS queue lives in. Env: AWS_DEFAULT_REGION. "
        "Default: us-east-1.",
    )
    args = parser.parse_args(argv)

    queue_url = args.queue_url or os.environ.get("SCRAPE_JOBS_QUEUE_URL")
    laravel_base_url = args.laravel_base_url or os.environ.get(
        "LARAVEL_BASE_URL", "http://host.docker.internal"
    )
    worker_token = args.worker_token or os.environ.get("LARAVEL_WORKER_TOKEN")
    aws_region = args.aws_region or os.environ.get("AWS_DEFAULT_REGION", "us-east-1")

    if not queue_url:
        parser.error("--queue-url is required (or set SCRAPE_JOBS_QUEUE_URL)")
    if not worker_token:
        parser.error("--worker-token is required (or set LARAVEL_WORKER_TOKEN)")

    return Config(
        queue_url=queue_url,
        laravel_base_url=laravel_base_url.rstrip("/"),
        worker_token=worker_token,
        aws_region=aws_region,
    )


def _report_result(
    config: Config, job_type: str, job_id: int, passthrough: dict, result: dict
) -> None:
    url = f"{config.laravel_base_url}/api/scrape-jobs/{job_type}/{job_id}/complete"
    resp = requests.post(
        url,
        json={**passthrough, **result},
        headers={"Authorization": f"Bearer {config.worker_token}"},
        timeout=15,
    )
    resp.raise_for_status()


def _handle_message(config: Config, sqs, message: dict) -> None:
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
        _report_result(config, job_type, job_id, passthrough, result)
    except Exception:
        logger.exception(
            "job failed (type=%s id=%s) — leaving message for SQS redrive/DLQ",
            job_type,
            job_id,
        )
        return

    sqs.delete_message(QueueUrl=config.queue_url, ReceiptHandle=receipt_handle)
    logger.info("job complete: type=%s id=%s", job_type, job_id)


def main(argv: list[str] | None = None) -> None:
    config = _parse_config(argv)
    sqs = boto3.client("sqs", region_name=config.aws_region)

    logger.info(
        "worker starting, polling %s, reporting to %s",
        config.queue_url,
        config.laravel_base_url,
    )
    while True:
        try:
            resp = sqs.receive_message(
                QueueUrl=config.queue_url,
                MaxNumberOfMessages=1,
                WaitTimeSeconds=20,
            )
        except Exception:
            logger.exception("SQS receive_message failed, backing off 5s")
            time.sleep(5)
            continue

        for message in resp.get("Messages", []):
            _handle_message(config, sqs, message)


if __name__ == "__main__":
    main()
