"""AWS Lambda handler: fast single-page capture (HTML + full-page
screenshot) to S3 -- the SeleniumBase (Stealthy Playwright Mode) counterpart
of ../grab/lambda_handler.py.

Given a single URL, this handler:
  1. Navigates to the page (optionally through a proxy).
  2. Captures the rendered HTML and a full-page screenshot.
  3. Uploads both to S3, concurrently.

Deliberately minimal -- no link/meta extraction, no consent-dialog handling,
no retry orchestration, no headed/Xvfb mode. Every step that isn't required
for "visit a URL, save its content and a screenshot" is left out on purpose,
same as ../grab.

Uses the same **sync** Playwright API as ../grab (`page.goto`, `page.content`,
`page.screenshot`, ...) -- the only difference from ../grab is *how* the
stealthy browser is launched: `seleniumbase_playwright.launch()` (baked into
../../images/seleniumbase-base) instead of `cloakbrowser.launch()`. Both
return an object exposing `.new_page()` / `.close()` and hand back a normal
`playwright.sync_api.Page`, so every line below this import is unchanged
from ../grab. See ../../images/seleniumbase-base/seleniumbase_playwright.py
for how that provider swap is bridged, and for the handful of interface
differences (`humanize`/`human_preset`/`geoip` are accepted below for
call-site parity with ../grab, but are no-ops with this provider -- see that
module's docstring for why).

Event schema:
    url                   str        required -- the page to capture (http/https only)
    proxy                 str|dict   none -- "http://user:pass@host:port" or a
                                     Playwright ProxySettings dict
                                     (credentials are split out automatically --
                                     see _normalize_proxy)
    geoip                 bool       false -- accepted for parity with ../grab;
                                     NO-OP with this provider (SeleniumBase has no
                                     built-in geoip-based timezone/locale
                                     derivation -- pass `timezone=`/`locale=`
                                     directly if you need specific values)
    bucket                str        "bytelyon-private"
    prefix                str        "grabs/{aws_request_id}/" -- S3 key prefix
    headless              bool       true
    humanize              bool       false -- accepted for parity with ../grab;
                                     NO-OP with this provider (see module
                                     docstring in seleniumbase_playwright.py)
    human_preset          str        "careful" -- accepted for parity; NO-OP
    wait_until            str        "domcontentloaded" -- passed to page.goto
                                     ("load"|"domcontentloaded"|"networkidle"|"commit")
    goto_timeout_ms        int        30000
    full_page_screenshot   bool       true

Returns:
    {
      "url": "https://example.com/",
      "content_key": "grabs/<request-id>/example_com-<hash>.html",
      "screenshot_key": "grabs/<request-id>/example_com-<hash>.png"
    }

Lambda-specific additions (see ../../images/seleniumbase-base/INSTRUCTIONS.md
for the same rationale in the base image, and ../grab/lambda_handler.py for
the CloakBrowser-based version of the same additions):
  - `--no-zygote` Chromium flag (`--disable-dev-shm-usage` is already one of
    SeleniumBase's own default CDP Mode browser args, and `--no-sandbox` is
    always applied by seleniumbase_playwright.launch() -- neither needs to
    be passed here).
  - URL scheme + SSRF validation before *and* after navigation (defense in
    depth against private/internal network access).
  - A TCP preflight on the proxy (if any) so an unreachable/blocking proxy
    fails in ~8s instead of hanging for the full function timeout (same as
    ../grab and ../serp-scraper).

Speed choices (identical rationale to ../grab):
  - `humanize`/`human_preset` are no-ops here (see above), but harmless to
    pass -- no need to special-case handler code per stealth provider.
  - `wait_until` defaults to `"domcontentloaded"`, which fires as soon as
    the DOM is parsed rather than waiting for every subresource (images,
    fonts, analytics beacons) to finish loading.
  - No link/meta extraction, no consent-dialog handling, no retry
    orchestration.
  - HTML and screenshot are uploaded to S3 concurrently (two independent
    network calls), not sequentially.
"""

from __future__ import annotations

import hashlib
import ipaddress
import logging
import re
import socket
import time
from concurrent.futures import ThreadPoolExecutor
from typing import Any
from urllib.parse import unquote, urlparse

import boto3
from playwright.sync_api import Page, ProxySettings
from playwright.sync_api import TimeoutError as PlaywrightTimeoutError
from seleniumbase_playwright import launch

logger = logging.getLogger("seleniumbase.lambda")
logger.setLevel(logging.INFO)

# awslambdaric wires the root logger with a formatter like
# "[INFO]\t2024-01-15T10:30:45.123Z\t<uuid>\tmessage" -- CloudWatch already
# timestamps every event server-side, so repeating timestamp + request id on
# every line just burns terminal columns. Trim to level + message.
for _h in logging.getLogger().handlers:
    _h.setFormatter(logging.Formatter("%(levelname)s %(message)s"))

DEFAULT_BUCKET = "bytelyon-private"

# How long to wait for a raw TCP handshake with the proxy before giving up.
# `seleniumbase_playwright.launch(proxy=...)` has no timeout of its own -- if
# the proxy silently drops the connection (common for residential-proxy
# providers that block/throttle datacenter-origin traffic, e.g. AWS Lambda's
# egress IPs) the whole invocation would otherwise hang until Lambda's
# function timeout, burning the full billed duration with no diagnostic
# signal. This preflight fails fast with a clear, actionable error instead.
# Same as ../grab and ../serp-scraper.
_PROXY_PREFLIGHT_TIMEOUT_S = 8.0


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


def _normalize_proxy(proxy: Any) -> ProxySettings | None:
    """Normalize a `proxy` event value into Playwright's `ProxySettings`
    shape: {"server": "<scheme>://<host>:<port>", "bypass"?, "username"?,
    "password"?}.

    seleniumbase_playwright.launch() accepts this same shape (or a plain
    string) and converts it internally into the "[user:pass@]host:port"
    string `seleniumbase.sb_cdp.Chrome(proxy=...)` expects -- see that
    module's `_normalize_proxy()` docstring for why SeleniumBase's CDP Mode
    doesn't need Playwright's own proxy-auth handling. A dict already in
    this shape is passed through unchanged. Same logic as ../grab and
    ../serp-scraper.
    """
    if not proxy:
        return None
    if isinstance(proxy, dict):
        return proxy
    if not isinstance(proxy, str):
        raise ValueError(
            f"'proxy' must be a string or dict, got {type(proxy).__name__}"
        )

    parsed = urlparse(proxy)
    if not parsed.hostname:
        raise ValueError(f"'proxy' is not a valid proxy URL: {proxy!r}")

    server = f"{parsed.scheme or 'http'}://{parsed.hostname}"
    if parsed.port:
        server += f":{parsed.port}"

    normalized: ProxySettings = ProxySettings(server=server)
    if parsed.username:
        normalized["username"] = unquote(parsed.username)
    if parsed.password:
        normalized["password"] = unquote(parsed.password)
    return normalized


def _proxy_host_port(proxy: ProxySettings | None) -> tuple[str, int] | None:
    """Best-effort (host, port) extraction from an already-normalized proxy
    dict, for the TCP preflight below. Returns None (preflight skipped, not
    fatal) if it can't be determined.
    """
    if not proxy:
        return None
    server = proxy.get("server")
    if not isinstance(server, str) or not server:
        return None
    parsed = urlparse(server if "://" in server else f"http://{server}")
    if not parsed.hostname or not parsed.port:
        return None
    return parsed.hostname, parsed.port


def _preflight_proxy(
    proxy: ProxySettings | None, timeout: float = _PROXY_PREFLIGHT_TIMEOUT_S
) -> None:
    """Raw TCP reachability check only -- deliberately does not attempt proxy
    auth or a CONNECT tunnel, just confirms *something* is listening on
    host:port so an unreachable/blocking proxy fails in ~8s instead of
    hanging for the full function timeout. A pass here does not guarantee
    the proxy will actually authenticate or tunnel traffic.
    """
    host_port = _proxy_host_port(proxy)
    if host_port is None:
        return
    host, port = host_port
    started = time.monotonic()
    try:
        with socket.create_connection((host, port), timeout=timeout):
            pass
    except OSError as e:
        raise RuntimeError(
            f"Proxy preflight failed: could not open a TCP connection to "
            f"{host}:{port} within {timeout}s ({e}). This is a network-level "
            f"failure, not a credentials issue -- it usually means the proxy "
            f"provider is blocking or silently dropping connections from "
            f"this network (e.g. AWS Lambda's egress IP ranges). Verify with "
            f"the proxy provider whether cloud/datacenter source IPs are "
            f"allowed on this gateway."
        ) from e
    logger.info(
        "Proxy preflight OK: TCP connect to %s:%s succeeded in %.2fs",
        host,
        port,
        time.monotonic() - started,
    )


def _s3_key(prefix: str, url: str, ext: str) -> str:
    parsed = urlparse(url)
    path = parsed.path.strip("/") or "index"
    safe = re.sub(r"[^a-zA-Z0-9/_-]", "_", f"{parsed.netloc}/{path}")
    digest = hashlib.sha256(url.encode()).hexdigest()[:10]
    return f"{prefix.rstrip('/')}/{safe}-{digest}.{ext}"


def handler(event: dict, context: Any) -> dict:
    url = event.get("url")
    if not isinstance(url, str) or not url.strip():
        raise ValueError("event must include a non-empty 'url' string")
    _validate_url(url)

    bucket = event.get("bucket", DEFAULT_BUCKET)
    prefix = event.get(
        "prefix", f"grabs/{getattr(context, 'aws_request_id', 'local')}/"
    )

    s3 = boto3.client("s3")

    proxy: ProxySettings | None = _normalize_proxy(event.get("proxy"))
    _preflight_proxy(proxy)

    browser = launch(
        headless=event.get("headless", True),
        humanize=event.get("humanize", False),
        human_preset=event.get("human_preset", "careful"),
        proxy=proxy,
        # Accepted for call-site parity with ../grab -- currently a no-op
        # with this provider. See seleniumbase_playwright.py's docstring.
        geoip=event.get("geoip", False),
    )
    page: Page = browser.new_page()
    try:
        try:
            page.goto(
                url,
                wait_until=event.get("wait_until", "domcontentloaded"),
                timeout=event.get("goto_timeout_ms", 30_000),
            )
        except PlaywrightTimeoutError:
            # Best-effort: capture whatever loaded so far rather than fail
            # the whole invocation.
            pass

        _validate_url(page.url)  # re-check after redirects

        html = page.content()
        png = page.screenshot(full_page=event.get("full_page_screenshot", True))

        content_key = _s3_key(prefix, page.url, "html")
        screenshot_key = _s3_key(prefix, page.url, "png")

        # Two independent uploads -- run them concurrently instead of
        # back-to-back to shave the S3 round-trip time roughly in half.
        with ThreadPoolExecutor(max_workers=2) as pool:
            content_future = pool.submit(
                s3.put_object,
                Bucket=bucket,
                Key=content_key,
                Body=html.encode("utf-8"),
                ContentType="text/html; charset=utf-8",
            )
            screenshot_future = pool.submit(
                s3.put_object,
                Bucket=bucket,
                Key=screenshot_key,
                Body=png,
                ContentType="image/png",
            )
            content_future.result()
            screenshot_future.result()

        return {
            "url": page.url,
            "content_key": content_key,
            "screenshot_key": screenshot_key,
        }
    finally:
        browser.close()
