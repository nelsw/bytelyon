"""Google-SERP navigation handler — the local-machine equivalent of
../../lambda/serp-di/lambda_handler.py, trimmed to the generic worker
contract: no SERP parsing here anymore (see app/Support/Serp.php in the main
app), just launch/navigate/capture/upload, same as every other handler.

The launch()/goto()/type() sequence, the DataImpulse SOCKS5 proxy
credentials, and the `human_preset="careful"` launch option are copied
verbatim from the script that ran reliably against Google overnight — do
not "improve" or re-tune any of that without re-validating against a live
run first.

The DataImpulse proxy is only the *fallback*: if the bot that enqueued this
job has one or more proxies configured in Settings, Laravel's `BotJob`
picks one at random and rides it along as the job's `proxy` field (see
common.proxy_settings()), which takes priority over DataImpulse. Bots with
no proxy of their own keep using DataImpulse, since Google needs *some*
residential/trusted-geo egress to reliably avoid a block/CAPTCHA page.

Unlike the Lambda version this was derived from, this runs in a normal
Docker container (not AWS Lambda), so:
  - `geoip=True` is used directly. cloakbrowser's own httpx-based geoip
    resolution only hangs inside *Lambda's* sandboxed network namespace
    (confirmed empirically — see ../../lambda/serp-di/lambda_handler.py's
    `_resolve_geo` docstring for the full story); it works fine in an
    ordinary container. Still wrapped in a fallback to `geoip=False` in
    case a given proxy/exit-IP combination misbehaves, same degrade-
    gracefully pattern ../../lambda/serp-scraper uses.
  - The `--disable-dev-shm-usage --no-zygote` Chromium args are Lambda-
    specific hardening and technically unnecessary here, but harmless
    (Chromium ignores flags it doesn't need) — kept for parity with the
    Lambda version rather than maintaining two divergent arg lists.
"""

from __future__ import annotations

import logging

from cloakbrowser import ProxySettings, launch
from common import capture_and_upload, proxy_settings
from playwright.sync_api import Browser, Page
from playwright.sync_api import TimeoutError as PlaywrightTimeoutError

logger = logging.getLogger("worker.serp")

GOOGLE_URL = "https://www.google.com"
PROXY_USERNAME = "2b9bc9b578274c5e9c8f__cr.us"
PROXY_PASSWORD = "a1d0893a3f201adf"
PROXY_HOST = "gw.dataimpulse.com"
PROXY_PORT = 824

_CONSENT_SELECTORS = [
    "#L2AGLb",
    "button:has-text('Accept all')",
    "button:has-text('I agree')",
]

# DataImpulse's SOCKS5 endpoint has a known non-trivial raw connection-
# failure rate — the retry wraps the *entire* attempt (fresh browser, fresh
# proxy connection) rather than just launch(), since a fresh SOCKS5
# connection on the next attempt routinely succeeds where the previous one
# didn't, and that failure has shown up at every stage (launch, geoip,
# navigation) in testing.
SCRAPE_ATTEMPTS = 2


def _default_proxy() -> ProxySettings:
    return ProxySettings(
        username=PROXY_USERNAME,
        password=PROXY_PASSWORD,
        bypass=".google.com",
        server=f"socks5://{PROXY_HOST}:{PROXY_PORT}",
    )


def _launch(headless: bool, proxy: ProxySettings | None) -> Browser:
    # A bot-configured proxy (picked at random on the Laravel side when a
    # bot has more than one -- see common.proxy_settings()) always wins;
    # this DataImpulse SOCKS5 endpoint is only the fallback for bots with no
    # proxy of their own, since Google needs *some* residential/trusted-geo
    # egress to reliably avoid a block/CAPTCHA page.
    common_kwargs = dict(
        proxy=proxy or _default_proxy(),
        headless=headless,
        human_preset="careful",
        args=["--disable-dev-shm-usage", "--no-zygote"],
    )
    try:
        return launch(geoip=True, **common_kwargs)
    except RuntimeError as e:
        if "geoip" not in str(e).lower():
            raise
        logger.warning("geoip resolution failed (%s); retrying without geoip", e)
        return launch(geoip=False, **common_kwargs)


def _dismiss_consent(page: Page) -> None:
    for selector in _CONSENT_SELECTORS:
        try:
            locator = page.locator(selector).first
            if locator.is_visible(timeout=1_000):
                locator.click(timeout=2_000)
                page.wait_for_load_state("domcontentloaded", timeout=5_000)
                return
        except Exception:
            continue


def _attempt(query: str, headless: bool, proxy: ProxySettings | None) -> dict:
    browser = _launch(headless, proxy)
    try:
        page: Page = browser.new_page()
        page.goto(GOOGLE_URL, timeout=45_000)
        page.locator('textarea[name="q"]').first.click()
        page.keyboard.type(query)
        page.keyboard.press("Enter")

        # A soft wait, not a hard requirement: Google's CAPTCHA/interstitial
        # page (google.com/sorry/...) never has #sfooter, so a bare
        # wait_for_selector would throw before a single byte gets captured.
        # Whether it showed up or not, capture unconditionally below — the
        # main app decides "blocked or not" once it has the HTML in hand
        # (see app/Support/Serp.php::blocked()), not this worker.
        try:
            page.wait_for_selector("#sfooter", timeout=30_000)
            _dismiss_consent(page)
        except PlaywrightTimeoutError:
            pass

        return capture_and_upload(page, job_type="serp")
    finally:
        browser.close()


def run(job: dict) -> dict:
    query = job["query"]
    headless = job.get("headless", False)
    proxy = proxy_settings(job)
    last_err: Exception | None = None
    for attempt in range(1, SCRAPE_ATTEMPTS + 1):
        try:
            return _attempt(query, headless, proxy)
        except Exception as e:
            last_err = e
            logger.warning("serp attempt %d/%d failed: %s", attempt, SCRAPE_ATTEMPTS, e)
    raise last_err  # type: ignore[misc]
