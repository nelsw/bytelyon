"""AWS Lambda handler: warm up the shared CloakBrowser base image.

Sole purpose: launch (and immediately close) a CloakBrowser browser instance
on every invocation. That launch is the expensive part of a cold start every
other handler in this repo also pays for on its own first invocation —
extracting/verifying the baked Chromium binary, spawning the
renderer/GPU-process tree, injecting CloakBrowser's stealth patches — but
this handler does no navigation, no scraping, and no S3/network I/O of its
own. Point a periodic trigger (e.g. an EventBridge scheduled rule) at this
function to keep its execution environment warm, or invoke it once after a
deploy as a cheap smoke test that the base image still boots correctly.

IMPORTANT — this does NOT warm any *other* Lambda function's execution
environment. Even functions built from the exact same container image get
their own independent pool of warm containers; invoking this function has no
guaranteed effect on ../news-scraper, ../page-scraper, ../scraper,
../serp-scraper, or ../site-crawler. Deploy this as its own function when you
specifically want a side-effect-free, zero-extra-IAM-permission way to keep
*a* CloakBrowser container warm and/or verify the base image boots — if your
goal is instead to stop a *specific* handler from going cold, the more
direct fix is either AWS Lambda Provisioned Concurrency on that handler's own
function, or a scheduled ping of that handler's own function with a cheap
real event.

Same sibling of ../news-scraper, ../page-scraper, ../scraper,
../serp-scraper, and ../site-crawler — built FROM the shared ../base-image.

Event schema:
    headless   bool   true
    navigate   bool   false — additionally open a page and navigate it to
                      "about:blank" (no network call) to also exercise
                      Chromium's page/render pipeline, not just process
                      startup

Returns:
    {
      "warm": true,
      "cold_start": true,             -- whether this was this container's
                                          first invocation since it started;
                                          expect "false" on later pings if
                                          the warm-up trigger is working
      "cloakbrowser_version": "0.5.10",
      "chromium_version": "146.0.7680.177.5",
      "launch_ms": 812.4,
      "navigate_ms": null,            -- null unless "navigate": true
      "total_ms": 815.9
    }
"""

from __future__ import annotations

import time
from typing import Any

import cloakbrowser  # pyright: ignore[reportMissingImports]
from cloakbrowser import launch  # pyright: ignore[reportMissingImports]
from playwright.sync_api import Page  # pyright: ignore[reportMissingImports]

# Module-level state persists for the life of the container (i.e. across
# every invocation it serves after the first), same as any other global in a
# Lambda handler module. Flips to True on this container's very first
# invocation, so every later invocation on the same (warm) container reports
# cold_start=False — a simple, honest signal that the warm-up trigger is
# actually keeping this function's containers alive.
_warmed = False


def handler(event: dict, context: Any) -> dict:
    global _warmed
    cold_start = not _warmed
    _warmed = True

    started = time.monotonic()

    browser = launch(
        headless=event.get("headless", True),
        args=[
            # Lambda /dev/shm is ~64 MB — Chromium crashes mid-render without this.
            "--disable-dev-shm-usage",
            # Lambda's restricted process model can't fork from Chromium's
            # zygote — without this, child renderer processes fail to spawn.
            "--no-zygote",
        ],
    )
    launch_done = time.monotonic()

    navigate_ms: float | None = None
    try:
        page: Page = browser.new_page()
        if event.get("navigate", False):
            nav_started = time.monotonic()
            page.goto("about:blank")
            navigate_ms = (time.monotonic() - nav_started) * 1000
    finally:
        browser.close()

    total_ms = (time.monotonic() - started) * 1000

    return {
        "warm": True,
        "cold_start": cold_start,
        "cloakbrowser_version": cloakbrowser.__version__,
        "chromium_version": cloakbrowser.CHROMIUM_VERSION,
        "launch_ms": round((launch_done - started) * 1000, 1),
        "navigate_ms": round(navigate_ms, 1) if navigate_ms is not None else None,
        "total_ms": round(total_ms, 1),
    }
