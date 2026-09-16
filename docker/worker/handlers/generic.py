"""Generic single-URL page capture — used for `news` and `sitemap` jobs (and
any future job type that's just "go to this URL and save it"). Local-worker
equivalent of ../../lambda/grab/lambda_handler.py's core logic, trimmed to
this worker fleet's shared contract.

Proxy optional: news/sitemap targets are ordinary websites, not Google —
nowhere near as aggressive about datacenter-vs-residential IP reputation,
and this worker's home-network egress is already more trustworthy than
Lambda's own IPs were anyway. So unlike handlers/serp.py, there's no
default proxy here — jobs run proxy-less unless the bot that enqueued them
had one or more proxies configured in Settings, in which case Laravel's
`BotJob` picks one at random and rides it along as the job's `proxy` field
(see common.proxy_settings()).

Two interchangeable browser providers, both exposing the same
`launch(...) -> Browser`-with-`.new_page()`/`.close()` interface:
  - "seleniumbase" (default) — SeleniumBase's Stealthy Playwright Mode (see
    ../../images/seleniumbase-base/seleniumbase_playwright.py). Not as
    battle-tested against Google specifically as CloakBrowser, but plenty
    stealthy for ordinary news/sitemap targets, and — the actual point of
    defaulting to it here — doesn't consume one of CloakBrowser Pro's 5
    concurrent-session seats. Those are reserved for handlers/serp.py,
    which actually needs CloakBrowser's Google-specific patches; news and
    sitemap jobs don't, and can run as many of these concurrently as the
    machine can handle without ever touching the CloakBrowser seat pool.
  - "cloakbrowser" — same provider handlers/serp.py uses.
Override the *starting* provider per-job with a `provider` field in the SQS
message, or fleet-wide via the GENERIC_BROWSER_PROVIDER env var.

Automatic fallback: if the starting attempt raises (launch/navigation
failure) or comes back with a block-shaped HTTP status (403/429/503), this
retries once with "cloakbrowser" — never the other direction, since
cloakbrowser is already the stronger/more expensive option and there's
nothing left to escalate to if *it* fails. This only escalates on a real
signal (an exception, or a status code that specifically means "you got
blocked/rate-limited"), not merely "the page looked different than
expected" — this handler has no per-site knowledge to judge that with, so
it doesn't try to. A `page.goto` timeout is still treated as best-effort
(capture whatever loaded rather than fail/escalate) — same as before —
since plenty of ordinary slow pages hit that without being blocked at all.
"""

from __future__ import annotations

import logging
import os

from cloakbrowser import launch as _cloakbrowser_launch
from common import capture_and_upload, proxy_settings
from playwright.sync_api import Page
from playwright.sync_api import TimeoutError as PlaywrightTimeoutError
from seleniumbase_playwright import launch as _seleniumbase_launch

logger = logging.getLogger("worker.generic")

_LAUNCHERS = {
    "cloakbrowser": _cloakbrowser_launch,
    "seleniumbase": _seleniumbase_launch,
}
DEFAULT_PROVIDER = (
    os.environ.get("GENERIC_BROWSER_PROVIDER", "seleniumbase").strip().lower()
)

# HTTP statuses that specifically mean "you got blocked/rate-limited", not
# just "this page doesn't exist" (404) or "the site is broken" (500) —
# escalating the provider wouldn't help with those, only with statuses a
# stealthier browser might actually get past.
_BLOCK_STATUS_CODES = {403, 429, 503}


def _attempt(
    url: str, job_type: str, headless: bool, provider: str, proxy: dict | None
) -> dict:
    launch = _LAUNCHERS[provider]
    browser = launch(
        headless=headless,
        proxy=proxy,
        args=["--disable-dev-shm-usage", "--no-zygote"],
    )
    try:
        page: Page = browser.new_page()
        response = None
        try:
            response = page.goto(url, wait_until="domcontentloaded", timeout=30_000)
        except PlaywrightTimeoutError:
            # Best-effort: capture whatever loaded so far rather than fail
            # the whole job. Plenty of ordinary slow pages hit this without
            # being blocked at all, so this alone doesn't trigger fallback.
            pass

        if response is not None and response.status in _BLOCK_STATUS_CODES:
            raise RuntimeError(f"got HTTP {response.status} from {url} via {provider}")

        return capture_and_upload(page, job_type=job_type)
    finally:
        browser.close()


def run(job: dict) -> dict:
    url = job["url"]
    job_type = job["type"]
    headless = job.get("headless", True)
    provider = str(job.get("provider", DEFAULT_PROVIDER)).strip().lower()
    if provider not in _LAUNCHERS:
        provider = "seleniumbase"
    proxy = proxy_settings(job)

    try:
        return _attempt(url, job_type, headless, provider, proxy)
    except Exception as e:
        if provider == "cloakbrowser":
            # Already the strongest provider available — nothing left to
            # escalate to, let the caller's own retry/redrive handle it.
            raise
        logger.warning(
            "generic scrape via %s failed (%s); falling back to cloakbrowser for %s",
            provider,
            e,
            url,
        )
        return _attempt(url, job_type, headless, "cloakbrowser", proxy)
