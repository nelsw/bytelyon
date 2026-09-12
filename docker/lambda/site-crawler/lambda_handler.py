"""AWS Lambda handler: breadth-first site crawl with per-page screenshots
uploaded to S3.

Starting from one or more seed URLs, this handler:
  1. Visits a page with cloakbrowser's stealth Chromium.
  2. Screenshots it and uploads the screenshot to S3.
  3. Extracts same-site links from the page and enqueues unvisited ones.
  4. Repeats breadth-first until `max_pages`/`max_depth` is hit, the crawl
     runs out of same-site links, or the Lambda invocation is running low on
     time.

Returns the list of URLs actually crawled (in visit order) — screenshots are
a side effect written to S3, not part of the return value.

Uses the **sync** cloakbrowser/Playwright API (`cloakbrowser.launch`), same
as ../news-scraper — this is fine under awslambdaric, which
invokes `handler(event, context)` synchronously with no asyncio event loop
already running.

Event schema:
    url                              str        a single seed URL
    urls                             list[str]  one or more seed URLs (checked before `url`)
                                                 Exactly one of `url` / `urls` is required.
    bucket                           str        "bytelyon-private"
    prefix                           str        "crawls/{aws_request_id}/" — S3 key prefix
    max_pages                        int        20  — hard cap on pages visited across all seeds
    max_depth                        int         2  — link-following depth from each seed (0 = seeds only)
    same_domain_only                 bool     true  — only follow links whose hostname matches the seed's
    full_page_screenshot             bool     true  — capture the entire scrollable page
    headless                         bool     true  — matches ../news-scraper
    humanize                         bool     true
    human_preset                     str    "careful"
    goto_timeout_ms                  int     30000
    wait_for_load_state_timeout_ms   int      5000

Returns:
    ["https://example.com/", "https://example.com/about", ...]

S3 layout:
    s3://{bucket}/{prefix}{sanitized-host-and-path}-{sha256[:10]}.png

Lambda-specific additions (see ../aws_lambda/INSTRUCTIONS.md for the same
rationale in the sibling handlers):
  - `--disable-dev-shm-usage` / `--no-zygote` Chromium flags.
  - URL scheme + SSRF validation before *and* after navigating to every URL —
    seeds and, crucially, every link discovered on a crawled page, since
    those are effectively attacker-influenced input if the crawled site is
    untrusted.
  - A remaining-time check (`context.get_remaining_time_in_millis()`) stops
    enqueuing new pages once the invocation is close to its timeout, so the
    function returns whatever it crawled so far instead of being killed
    mid-request.
"""

from __future__ import annotations

import hashlib
import ipaddress
import logging
import re
import socket
from collections import deque
from typing import Any
from urllib.parse import urldefrag, urlparse

import boto3
from cloakbrowser import launch
from playwright.sync_api import Page
from playwright.sync_api import TimeoutError as PlaywrightTimeoutError

logger = logging.getLogger("cloakbrowser.lambda")
logger.setLevel(logging.INFO)

# awslambdaric wires the root logger with a formatter like
# "[INFO]\t2024-01-15T10:30:45.123Z\t<uuid>\tmessage" — CloudWatch already
# timestamps every event server-side, so repeating timestamp + request id on
# every line just burns terminal columns. Trim to level + message.
for _h in logging.getLogger().handlers:
    _h.setFormatter(logging.Formatter("%(levelname)s %(message)s"))

DEFAULT_BUCKET = "bytelyon-private"
_TIME_BUDGET_MS = 10_000  # stop starting new pages this close to the Lambda deadline


def _validate_url(url: str) -> None:
    """Reject non-HTTP schemes and URLs that resolve to private/internal IPs."""
    parsed = urlparse(url)
    if parsed.scheme.lower() not in ("http", "https"):
        raise ValueError(
            f"Only http:// and https:// URLs are supported, got: {parsed.scheme!r}"
        )
    hostname = parsed.hostname
    if not hostname:
        raise ValueError("URL has no hostname")
    try:
        infos = socket.getaddrinfo(hostname, None, socket.AF_UNSPEC, socket.SOCK_STREAM)
    except socket.gaierror:
        raise ValueError(f"Cannot resolve hostname: {hostname}")
    for info in infos:
        addr = ipaddress.ip_address(info[4][0])
        if not addr.is_global:
            raise ValueError("URLs targeting private/internal networks are blocked")


def _normalize(url: str) -> str:
    """Strip the fragment so '#section' anchors don't count as distinct pages."""
    return urldefrag(url)[0]


def _same_site(seed_host: str, candidate_host: str) -> bool:
    return seed_host.lower() == candidate_host.lower()


def _s3_key(prefix: str, url: str) -> str:
    parsed = urlparse(url)
    path = parsed.path.strip("/") or "index"
    safe = re.sub(r"[^a-zA-Z0-9/_-]", "_", f"{parsed.netloc}/{path}")
    digest = hashlib.sha256(url.encode()).hexdigest()[:10]
    return f"{prefix.rstrip('/')}/{safe}-{digest}.png"


def _extract_links(page: Page) -> list[str]:
    try:
        return page.eval_on_selector_all("a[href]", "els => els.map(e => e.href)")
    except Exception as e:
        logger.warning("link extraction failed on %s: %s", page.url, e)
        return []


def _seeds_from_event(event: dict) -> list[str]:
    urls = event.get("urls")
    if urls is not None:
        if (
            not isinstance(urls, list)
            or not urls
            or not all(isinstance(u, str) for u in urls)
        ):
            raise ValueError("'urls' must be a non-empty list of strings")
        return urls
    url = event.get("url")
    if isinstance(url, str) and url:
        return [url]
    raise ValueError("event must include 'url' (str) or 'urls' (list[str])")


def _remaining_ms(context: Any) -> int:
    try:
        return context.get_remaining_time_in_millis()
    except Exception:
        return 10**9  # local RIE testing has no deadline — treat as unlimited


def handler(event: dict, context: Any) -> list[str]:
    seeds = _seeds_from_event(event)
    for seed in seeds:
        _validate_url(seed)

    bucket = event.get("bucket", DEFAULT_BUCKET)
    prefix = event.get(
        "prefix", f"crawls/{getattr(context, 'aws_request_id', 'local')}/"
    )
    max_pages = int(event.get("max_pages", 20))
    max_depth = int(event.get("max_depth", 2))
    same_domain_only = event.get("same_domain_only", True)
    full_page_screenshot = event.get("full_page_screenshot", True)

    s3 = boto3.client("s3")

    browser = launch(
        headless=event.get("headless", True),
        humanize=event.get("humanize", True),
        human_preset=event.get("human_preset", "careful"),
        args=[
            # Lambda /dev/shm is ~64 MB — Chromium crashes mid-render without this.
            "--disable-dev-shm-usage",
            # Lambda's restricted process model can't fork from Chromium's
            # zygote — without this, child renderer processes fail to spawn.
            "--no-zygote",
        ],
    )
    page: Page = browser.new_page()

    visited: set[str] = set()
    crawled: list[str] = []
    queue: deque[tuple[str, int, str]] = deque()  # (url, depth, seed_host)
    for seed in seeds:
        seed_host = urlparse(seed).hostname or ""
        queue.append((_normalize(seed), 0, seed_host))

    try:
        while queue and len(crawled) < max_pages:
            if _remaining_ms(context) < _TIME_BUDGET_MS:
                logger.warning(
                    "stopping crawl early: %d ms remaining, %d/%d pages done",
                    _remaining_ms(context),
                    len(crawled),
                    max_pages,
                )
                break

            url, depth, seed_host = queue.popleft()
            if url in visited:
                continue
            visited.add(url)

            try:
                _validate_url(url)
                page.goto(url, timeout=event.get("goto_timeout_ms", 30_000))
                page.wait_for_load_state(
                    "domcontentloaded",
                    timeout=event.get("wait_for_load_state_timeout_ms", 5_000),
                )
            except PlaywrightTimeoutError:
                # Best-effort: screenshot/links from whatever loaded so far.
                pass
            except Exception as e:
                logger.warning("skipping %s: %s", url, e)
                continue

            try:
                _validate_url(page.url)  # re-check after redirects
            except Exception as e:
                logger.warning("skipping %s after redirect: %s", url, e)
                continue

            final_url = _normalize(page.url)
            if final_url in visited:
                # Reached a URL already crawled under a different alias (e.g.
                # the seed lacked a trailing slash but resolved to one, or a
                # redirect landed on an already-visited page).
                continue
            visited.add(final_url)

            try:
                png = page.screenshot(full_page=full_page_screenshot)
                s3.put_object(
                    Bucket=bucket,
                    Key=_s3_key(prefix, final_url),
                    Body=png,
                    ContentType="image/png",
                )
            except Exception as e:
                logger.warning("screenshot/upload failed for %s: %s", final_url, e)

            crawled.append(final_url)

            if depth < max_depth:
                for link in _extract_links(page):
                    link = _normalize(link)
                    if link in visited:
                        continue
                    link_host = urlparse(link).hostname or ""
                    if same_domain_only and not _same_site(seed_host, link_host):
                        continue
                    if urlparse(link).scheme.lower() not in ("http", "https"):
                        continue
                    queue.append((link, depth + 1, seed_host))
    finally:
        browser.close()

    return crawled
