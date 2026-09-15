"""Generic single-URL page capture — used for `news` and `sitemap` jobs (and
any future job type that's just "go to this URL and save it"). Local-worker
equivalent of ../../lambda/grab/lambda_handler.py's core logic, trimmed to
this worker fleet's shared contract.

No proxy: news/sitemap targets are ordinary websites, not Google — nowhere
near as aggressive about datacenter-vs-residential IP reputation, and this
worker's home-network egress is already more trustworthy than Lambda's own
IPs were anyway. Add a proxy here later if a specific target needs it.
"""

from __future__ import annotations

from cloakbrowser import launch
from common import capture_and_upload
from playwright.sync_api import Page
from playwright.sync_api import TimeoutError as PlaywrightTimeoutError


def run(job: dict) -> dict:
    url = job["url"]
    job_type = job["type"]

    browser = launch(
        headless=job.get("headless", True),
        args=["--disable-dev-shm-usage", "--no-zygote"],
    )
    try:
        page: Page = browser.new_page()
        try:
            page.goto(url, wait_until="domcontentloaded", timeout=30_000)
        except PlaywrightTimeoutError:
            # Best-effort: capture whatever loaded so far rather than fail
            # the whole job.
            pass

        return capture_and_upload(page, job_type=job_type)
    finally:
        browser.close()
