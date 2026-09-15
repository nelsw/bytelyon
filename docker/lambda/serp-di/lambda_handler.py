"""AWS Lambda handler / SQS worker core wrapping — as closely as possible —
the CloakBrowser script that ran reliably against Google all night as a
standalone `uv run` script:

    #!/usr/bin/env -S uv run --script
    #
    # /// script
    # requires-python = ">=3.14"
    # dependencies = [
    #   "cloakbrowser",
    #   "cloakbrowser[geoip]",
    # ]
    # ///
    import argparse
    from pathlib import Path
    from uuid import uuid5, NAMESPACE_URL

    from cloakbrowser import launch, ProxySettings
    from playwright.sync_api import Browser, Page

    GOOGLE_URL = "https://www.google.com"

    def main(query: str, path: str, headless: bool) -> None:
        Path(path).mkdir(parents=True, exist_ok=True)
        browser: Browser = launch(
            proxy=ProxySettings(
                username='2b9bc9b578274c5e9c8f__cr.us',
                password='a1d0893a3f201adf',
                bypass='.google.com',
                server='socks5://gw.dataimpulse.com:824',
            ),
            headless=headless,
            human_preset="careful",
            geoip=True
        )
        page: Page = browser.new_page()
        page.goto(GOOGLE_URL)
        page.locator('textarea[name="q"]').first.click()
        page.keyboard.type(query)
        page.keyboard.press('Enter')
        page.wait_for_selector('#sfooter')
        name = uuid5(NAMESPACE_URL, f"{GOOGLE_URL}?q={query.replace(' ', '+')}")
        page.screenshot(path=f"{path}/{name}.png", full_page=True)
        with open(f"{path}/{name}.html", "w", encoding="utf-8") as file:
            file.write(page.content())
        browser.close()

The launch()/goto()/type() sequence, the DataImpulse SOCKS5 proxy
credentials, and the `human_preset="careful"` launch option are copied
verbatim — do not "improve" or re-tune any of that without re-validating
against a live run first, since this exact combination is the one that's
already proven reliable.

This module is used two ways:
  1. As an AWS Lambda handler (`handler(event, context)`), same as every
     sibling handler in ../serp-scraper etc.
  2. Imported directly by worker.py's SQS polling loop, which runs this same
     browser logic on a local machine instead of in Lambda — see
     ../serp-di/worker.py and its own docstring for why (Google trusts
     residential/home-network egress far more than any datacenter IP,
     Lambda's included).

Changes relative to the original script above:
  1. `argparse` -> function args / event fields (`query`, `headless`).
  2. The capture step (screenshot + HTML) no longer sits behind a bare
     `page.wait_for_selector("#sfooter")`. Confirmed in testing: Google's
     CAPTCHA/interstitial page (served from google.com/sorry/...) never has
     that selector, so a block was indistinguishable from a crash — the
     handler raised a TimeoutError before capturing a single byte, with no
     screenshot or HTML to diagnose *why*. The wait is now soft-bounded, and
     capture always runs afterward regardless of that outcome.
  3. Screenshot/HTML are captured in memory (`page.screenshot()` /
     `page.content()`) and uploaded straight to S3, then parsed into the same
     structured `data` shape (organic results, ads, PAA/similar queries,
     etc.) that ../serp-scraper's Lambda handler produces — see
     `_extract_data` below, ported from there — so this handler's output is
     a drop-in match for `Serp::update()`'s `data`/`screenshot_key`/
     `content_key` contract regardless of which handler (or which machine)
     produced it. No local disk round-trip needed anymore.
  4. GeoIP: `geoip=True` is NOT used — see `_resolve_geo()`'s docstring for
     why (it hangs specifically inside AWS Lambda's network namespace).
     Timezone/locale are resolved independently instead. Harmless overhead
     when run outside Lambda (worker.py), just an extra IP lookup.
  5. Lambda-only Chromium hardening (`--disable-dev-shm-usage --no-zygote`)
     is always applied. It's a no-op outside Lambda (worker.py) — Chromium
     ignores flags it doesn't need — so there's no reason to special-case it
     away for the local-worker use case.

Function args (`main`) / event schema (`handler`):
    query       str   required, e.g. "sailing blocks"
    headless    bool  optional, default True
    bucket      str   optional, default "bytelyon-private" — S3 bucket for
                       page content + screenshot uploads
    prefix      str   optional, default "serp-scrapes/di/<uuid4>/" — S3 key
                       prefix for both uploads

Returns:
    {
      "query": "...",
      "url": "https://www.google.com/search?q=...",
      "blocked": bool,
      "screenshot_key": "serp-scrapes/di/<uuid>/sailing_blocks-<hash>.png",
      "content_key": "serp-scrapes/di/<uuid>/sailing_blocks-<hash>.html",
      "data": {
        "sponsored_results": [...], "sponsored_products": [...],
        "organic_results": [...], "organic_products": [...],
        "similar_queries": [...]
      }
    }

Same `data` shape (and same degrade-independently-to-`[]`-on-failure
behavior) as ../serp-scraper/lambda_handler.py — see that module's own
docstring for the full per-field shape of each section.
"""

from __future__ import annotations

import hashlib
import logging
import re
import time
import uuid
from typing import Any
from urllib.parse import urlparse

import boto3
from cloakbrowser import ProxySettings, launch
from cloakbrowser.geoip import COUNTRY_LOCALE_MAP
from playwright.sync_api import Browser, Page
from playwright.sync_api import TimeoutError as PlaywrightTimeoutError

logger = logging.getLogger("cloakbrowser.serp_di")
logger.setLevel(logging.INFO)

GOOGLE_URL = "https://www.google.com"
PROXY_USERNAME = "2b9bc9b578274c5e9c8f__cr.us"
PROXY_PASSWORD = "a1d0893a3f201adf"
PROXY_HOST = "gw.dataimpulse.com"
PROXY_PORT = 824
PROXY_URL = f"socks5://{PROXY_USERNAME}:{PROXY_PASSWORD}@{PROXY_HOST}:{PROXY_PORT}"
GEOIP_DB_PATH = "/root/.cloakbrowser/geoip/GeoLite2-City.mmdb"
DEFAULT_BUCKET = "bytelyon-private"

# DataImpulse's SOCKS5 endpoint has a known non-trivial raw connection-failure
# rate (confirmed separately via plain HTTP sampling against this same proxy).
# It's shown up here at every stage: GeoIP resolution timing out (the exit-IP
# lookup itself can't get through), launch-time "Target crashed", and
# page.goto() itself timing out mid-navigation. All are proxy-connection
# flakiness, not a code bug, so the retry wraps the *entire* attempt (fresh
# browser, fresh proxy connection, fresh everything) rather than just launch()
# — a fresh SOCKS5 connection on the next attempt routinely succeeds where the
# previous one didn't, but only if it's actually a fresh connection.
SCRAPE_ATTEMPTS = 2

_BLOCK_PHRASES = (
    "unusual traffic",
    "our systems have detected",
)

# Google's consent interstitial (shown for some EU-ish exit IPs regardless of
# hl/gl) — #L2AGLb ("I agree") has been a long-lived id for this button, but
# fall back to text matching since it's not guaranteed to still exist.
_CONSENT_SELECTORS = [
    "#L2AGLb",
    "button:has-text('Accept all')",
    "button:has-text('I agree')",
]


def _resolve_geo() -> tuple[str | None, str | None]:
    """Resolve (timezone, locale) for the proxy's exit IP ourselves, instead
    of using cloakbrowser's built-in geoip=True.

    Confirmed by direct diagnostic (bytelyon-http-test Lambda, plain
    `requests` + PySocks against this same SOCKS5 proxy): the proxy itself
    responds through AWS Lambda's network in under a second, every time.
    But cloakbrowser's own geoip=True path (httpx's SOCKS5 client) reliably
    hangs to its full timeout when run inside Lambda specifically — not
    reproducible locally in Docker, and not reproducible with a different
    HTTP client from the very same Lambda function. That points at an
    httpx/socksio incompatibility with Lambda's sandboxed network namespace,
    not a proxy problem. Reusing the same baked GeoLite2 DB and
    country->locale map cloakbrowser itself uses, just via `requests`
    (already proven reliable here) instead of httpx. Also used unmodified by
    worker.py (outside Lambda) — a plain, reliable lookup either way.
    """
    import geoip2.database
    import requests

    try:
        resp = requests.get(
            "https://api.ipify.org",
            proxies={"https": PROXY_URL},
            timeout=10,
        )
        resp.raise_for_status()
        ip = resp.text.strip()
    except Exception:
        return None, None

    try:
        with geoip2.database.Reader(GEOIP_DB_PATH) as reader:
            city = reader.city(ip)
            timezone = city.location.time_zone
            country = city.country.iso_code
            locale = COUNTRY_LOCALE_MAP.get(country) if country else None
            return timezone, locale
    except Exception:
        return None, None


def _launch(headless: bool) -> Browser:
    timezone, locale = _resolve_geo()
    return launch(
        proxy=ProxySettings(
            username=PROXY_USERNAME,
            password=PROXY_PASSWORD,
            bypass=".google.com",
            server=f"socks5://{PROXY_HOST}:{PROXY_PORT}",
        ),
        headless=headless,
        human_preset="careful",
        # geoip=True is deliberately NOT used here — see _resolve_geo()'s
        # docstring. timezone/locale are resolved ourselves instead, using a
        # client already confirmed reliable through this proxy from Lambda.
        # Falls back to no timezone/locale override (Chromium's own defaults)
        # if resolution fails, rather than blocking the whole launch on it.
        timezone=timezone,
        locale=locale,
        # Lambda-only hardening, harmless outside Lambda (worker.py) —
        # Chromium simply ignores flags it doesn't need there. Without these
        # in Lambda, `browser.new_page()` reliably raises "Target crashed":
        # Lambda's /dev/shm is ~64MB (Chromium's renderer needs more) and its
        # sandboxed process model can't fork from Chromium's zygote process.
        # Same fix CloakHQ's own official AWS Lambda example bakes in
        # unconditionally for this exact reason.
        args=["--disable-dev-shm-usage", "--no-zygote"],
    )


def _empty_data() -> dict:
    """All-empty `data` shape, used when the page is blocked or extraction
    fails outright, so callers can always rely on the same response shape.
    Identical to ../serp-scraper/lambda_handler.py's own `_empty_data`.
    """
    return {
        "sponsored_results": [],
        "sponsored_products": [],
        "organic_results": [],
        "organic_products": [],
        "similar_queries": [],
    }


# Ported verbatim from ../serp-scraper/lambda_handler.py — see that module's
# own comments for the full rationale (structural, best-effort parsing that
# degrades gracefully as Google's unversioned/obfuscated markup shifts).
_DATA_EXTRACT_JS = r"""
() => {
  const textOf = (el) => (el && el.textContent ? el.textContent.trim() : "");

  const sponsoredResults = Array.from(document.querySelectorAll("[data-pcu]")).map((e, i) => {
    const spans = e.querySelectorAll("span");
    const rawBrand = spans[1] ? textOf(spans[1]) : "";
    return {
      index: i,
      title: spans[0] ? textOf(spans[0]) : "",
      brand: rawBrand.split("https://")[0].trim(),
      href: e.getAttribute("href") || "",
    };
  });

  const organicProducts = Array.from(document.querySelectorAll("product-viewer-entrypoint")).map((e, i) => {
    const div = e.querySelector("div");
    const img = e.querySelector("img");
    let title = div ? (div.getAttribute("aria-label") || "") : "";
    if (!title && img) title = img.getAttribute("alt") || "";
    return {
      index: i,
      title,
      image: img ? (img.getAttribute("src") || "") : "",
    };
  });

  const sponsoredProducts = Array.from(document.querySelectorAll("[data-dtld]")).map((e, i) => {
    const div = e.querySelector("div.pla-unit-container");
    const a = div ? div.querySelector("a.pla-unit-img-container-link") : null;
    const titleEl = div ? div.querySelector("div[data-call_grow_wiz_event='true']") : null;
    const priceEl = div ? div.querySelector("div[aria-label]") : null;
    const brandEl = div ? div.querySelector("span[role='text']") : null;
    const img = a ? a.querySelector("img") : null;
    return {
      index: i,
      domain: e.getAttribute("data-dtld") || "",
      href: a ? (a.getAttribute("href") || "") : "",
      image: img ? (img.getAttribute("src") || "") : "",
      title: textOf(titleEl),
      price: priceEl ? (priceEl.getAttribute("aria-label") || "") : "",
      brand: textOf(brandEl),
    };
  });

  const organicResults = Array.from(document.querySelectorAll("h3[id]")).map((e, i) => {
    const parent = e.parentElement;
    return {
      index: i,
      title: textOf(e),
      href: parent ? (parent.getAttribute("href") || "") : "",
    };
  });

  const similarQueries = [];
  document.querySelectorAll("div[data-notify-expansion]").forEach((e) => {
    const q = e.getAttribute("data-q") || "";
    if (q.length > 4) similarQueries.push(q);
  });
  const botstuff = document.querySelector("div#botstuff");
  if (botstuff) {
    botstuff.querySelectorAll("a").forEach((a) => {
      const t = textOf(a);
      if (t.length > 4) similarQueries.push(t);
    });
  }

  return { sponsoredResults, organicProducts, sponsoredProducts, organicResults, similarQueries };
}
"""

# In-memory cache of resolved destination URLs, shared across invocations in
# the same warm process (Lambda container or long-running worker.py loop).
_URL_CACHE: dict[str, tuple[float, str]] = {}
_URL_CACHE_TTL_S = 1800  # 30 minutes


def _to_domain(url: str) -> str:
    host = urlparse(url).hostname or ""
    return host.removeprefix("www.")


def _final_url(page: Page, link: str, timeout_ms: int) -> str:
    """Resolve a (possibly relative/redirect) Google result link to its final
    destination URL by actually requesting it, through the same browser
    context (and therefore the same proxy/session) as the search itself.
    Falls back to the normalized-but-unresolved link on any failure.
    """
    if not link:
        return link
    if not link.startswith("https://www.google.com"):
        if not link.startswith("/"):
            link = f"/{link}"
        link = f"https://www.google.com{link}"

    cache_key = hashlib.sha1(link.encode()).hexdigest()
    now = time.monotonic()
    cached = _URL_CACHE.get(cache_key)
    if cached and cached[0] > now:
        return cached[1]

    try:
        final_url = page.request.get(link, timeout=timeout_ms).url
    except Exception:
        return link

    expires_at = now + _URL_CACHE_TTL_S
    _URL_CACHE[cache_key] = (expires_at, final_url)
    _URL_CACHE[hashlib.sha1(final_url.encode()).hexdigest()] = (expires_at, final_url)
    return final_url


def _extract_data(page: Page, url_resolve_timeout_ms: int) -> dict:
    """Best-effort structural SERP parser — ported verbatim from
    ../serp-scraper/lambda_handler.py. See that module for full rationale.
    """
    try:
        raw = page.evaluate(_DATA_EXTRACT_JS)
    except Exception:
        logger.exception("SERP data extraction failed")
        return _empty_data()

    def resolve(link: str) -> str:
        return _final_url(page, link, url_resolve_timeout_ms)

    sponsored_results: list[dict] = []
    try:
        for item in raw.get("sponsoredResults", []):
            url = resolve(item.get("href", ""))
            sponsored_results.append(
                {
                    "kind": "sponsored_result",
                    "index": item["index"],
                    "title": item.get("title", ""),
                    "brand": item.get("brand", ""),
                    "url": url,
                    "domain": _to_domain(url),
                }
            )
    except Exception:
        logger.exception("sponsored_results extraction failed")
        sponsored_results = []

    organic_products: list[dict] = []
    try:
        organic_products = [
            {
                "kind": "organic_product",
                "index": item["index"],
                "title": item.get("title", ""),
                "image": item.get("image", ""),
            }
            for item in raw.get("organicProducts", [])
        ]
    except Exception:
        logger.exception("organic_products extraction failed")
        organic_products = []

    sponsored_products: list[dict] = []
    try:
        for item in raw.get("sponsoredProducts", []):
            url = resolve(item.get("href", ""))
            sponsored_products.append(
                {
                    "kind": "sponsored_product",
                    "index": item["index"],
                    "domain": item.get("domain", ""),
                    "url": url,
                    "image": item.get("image", ""),
                    "title": item.get("title", ""),
                    "price": item.get("price", ""),
                    "brand": item.get("brand", ""),
                }
            )
    except Exception:
        logger.exception("sponsored_products extraction failed")
        sponsored_products = []

    organic_results: list[dict] = []
    try:
        for item in raw.get("organicResults", []):
            url = resolve(item.get("href", ""))
            organic_results.append(
                {
                    "kind": "organic_result",
                    "index": item["index"],
                    "title": item.get("title", ""),
                    "url": url,
                    "domain": _to_domain(url),
                }
            )
    except Exception:
        logger.exception("organic_results extraction failed")
        organic_results = []

    similar_queries: list[dict] = []
    try:
        similar_queries = [
            {"kind": "similar_query", "index": i, "value": value}
            for i, value in enumerate(raw.get("similarQueries", []))
        ]
    except Exception:
        logger.exception("similar_queries extraction failed")
        similar_queries = []

    return {
        "sponsored_results": sponsored_results,
        "sponsored_products": sponsored_products,
        "organic_results": organic_results,
        "organic_products": organic_products,
        "similar_queries": similar_queries,
    }


def _dismiss_consent(page: Page) -> None:
    """Best-effort click-through for Google's cookie-consent interstitial."""
    for selector in _CONSENT_SELECTORS:
        try:
            locator = page.locator(selector).first
            if locator.is_visible(timeout=1_000):
                locator.click(timeout=2_000)
                page.wait_for_load_state("domcontentloaded", timeout=5_000)
                return
        except Exception:
            continue


def _is_blocked(page: Page, html: str, sfooter_found: bool) -> bool:
    """Checks the already-captured page source rather than issuing a second,
    separate live DOM query — see ../serp-scraper/lambda_handler.py's
    `_is_blocked` for why that matters (a real CAPTCHA page can be missed by
    a live query that races the page settling). `sfooter_found` folds in
    this handler's own original wait_for_selector('#sfooter') signal: that
    selector only ever appears on a real, unblocked results page.
    """
    if "/sorry/" in page.url:
        return True
    if not sfooter_found:
        return True
    return any(phrase in html.lower() for phrase in _BLOCK_PHRASES)


def _s3_key(prefix: str, query: str, url: str, ext: str) -> str:
    safe_query = (
        re.sub(r"[^a-zA-Z0-9_-]+", "_", query.strip()).strip("_")[:80] or "query"
    )
    digest = hashlib.sha256(url.encode()).hexdigest()[:10]
    return f"{prefix.rstrip('/')}/{safe_query}-{digest}.{ext}"


def _attempt(query: str, headless: bool, bucket: str, prefix: str) -> dict:
    browser: Browser = _launch(headless)
    try:
        page: Page = browser.new_page()
        page.goto(GOOGLE_URL, timeout=45_000)
        page.locator('textarea[name="q"]').first.click()
        page.keyboard.type(query)
        page.keyboard.press("Enter")

        # #sfooter only ever appears on a real, unblocked results page.
        # Google's CAPTCHA/interstitial page (served from google.com/sorry/...)
        # never has it, so a naked `wait_for_selector` throws before a single
        # byte of the page is ever captured — the original script's biggest
        # blind spot: a block looked identical to a crash, with no
        # screenshot/HTML to diagnose it. Downgrade the wait to a soft check
        # (bounded timeout, exception caught) so the capture step below
        # always runs and always returns something to look at, whether
        # that's a clean SERP or a CAPTCHA page.
        try:
            page.wait_for_selector("#sfooter", timeout=30_000)
            sfooter_found = True
        except PlaywrightTimeoutError:
            sfooter_found = False

        if sfooter_found:
            _dismiss_consent(page)

        final_url = page.url
        html = page.content()
        png = page.screenshot(full_page=True)
        blocked = _is_blocked(page, html, sfooter_found)

        content_key = _s3_key(prefix, query, final_url, "html")
        screenshot_key = _s3_key(prefix, query, final_url, "png")

        s3 = boto3.client("s3")
        s3.put_object(
            Bucket=bucket,
            Key=content_key,
            Body=html.encode("utf-8"),
            ContentType="text/html; charset=utf-8",
        )
        s3.put_object(
            Bucket=bucket, Key=screenshot_key, Body=png, ContentType="image/png"
        )

        if blocked:
            logger.warning("Google served a block/CAPTCHA page for query=%r", query)
            data = _empty_data()
        else:
            data = _extract_data(page, url_resolve_timeout_ms=10_000)

        return {
            "query": query,
            "url": final_url,
            "blocked": blocked,
            "screenshot_key": screenshot_key,
            "content_key": content_key,
            "data": data,
        }
    finally:
        browser.close()


def main(
    query: str,
    headless: bool = False,
    bucket: str = DEFAULT_BUCKET,
    prefix: str | None = None,
) -> dict:
    if prefix is None:
        prefix = f"serp-scrapes/di/{uuid.uuid4()}/"
    last_err: Exception | None = None
    for attempt in range(1, SCRAPE_ATTEMPTS + 1):
        try:
            return _attempt(query, headless, bucket, prefix)
        except Exception as e:
            last_err = e
            logger.warning(
                "scrape attempt %d/%d failed: %s", attempt, SCRAPE_ATTEMPTS, e
            )
    raise last_err  # type: ignore[misc]


def handler(event: dict, context: Any) -> dict:
    query = event.get("query")
    if not isinstance(query, str) or not query.strip():
        raise ValueError("event must include a non-empty 'query' string")
    return main(
        query=query,
        # False (headed, via the base image's Xvfb entrypoint) matches both
        # ../serp-scraper's own default and CloakBrowser's own guidance that
        # headed mode meaningfully reduces blocks — "most blocks come from
        # missing residential proxy + geoip + headed mode."
        headless=event.get("headless", False),
        bucket=event.get("bucket", DEFAULT_BUCKET),
        prefix=event.get("prefix"),
    )
