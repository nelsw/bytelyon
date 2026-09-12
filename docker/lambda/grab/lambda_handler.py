"""AWS Lambda handler: fast single-page capture (HTML + full-page
screenshot) to S3.

Given a single URL, this handler:
  1. Navigates to the page (optionally through a proxy).
  2. Captures the rendered HTML and a full-page screenshot.
  3. Uploads both to S3, concurrently.

Deliberately minimal — no link/meta extraction, no consent-dialog handling,
no retry orchestration, no headed/Xvfb mode. Every step that isn't required
for "visit a URL, save its content and a screenshot" is left out on purpose:
see ../page-scraper (adds link/meta extraction), ../serp-scraper (adds
Google-specific extraction + block detection), and ../scraper (adds retry
orchestration + headed Xvfb mode) for handlers that trade some of this
speed for richer output.

Uses the **sync** cloakbrowser/Playwright API (`cloakbrowser.launch`), same
as ../page-scraper, ../news-scraper, ../serp-scraper, and ../site-crawler —
this avoids the async event-loop plumbing and headed/Xvfb startup cost that
../scraper pays for its extra features.

Event schema:
    url                   str        required — the page to capture (http/https only)
    proxy                 str|dict   none — "http://user:pass@host:port" or a
                                     cloakbrowser/Playwright ProxySettings dict
                                     (credentials are split out automatically —
                                     see _normalize_proxy)
    geoip                 bool       false — auto-derive timezone/locale from the
                                     proxy's exit IP (only meaningful with `proxy` set)
    bucket                str        "bytelyon-private"
    prefix                str        "grabs/{aws_request_id}/" — S3 key prefix
    headless              bool       true
    humanize              bool       false — off by default for speed; set true for
                                     sites that fingerprint mouse/keyboard/scroll behavior
    human_preset          str        "careful"
    wait_until            str        "domcontentloaded" — passed to page.goto
                                     ("load"|"domcontentloaded"|"networkidle"|"commit")
    goto_timeout_ms        int        30000
    full_page_screenshot   bool       true

Returns:
    {
      "url": "https://example.com/",
      "content_key": "grabs/<request-id>/example_com-<hash>.html",
      "screenshot_key": "grabs/<request-id>/example_com-<hash>.png"
    }

Lambda-specific additions (see ../base-image/INSTRUCTIONS.md for the same
rationale in the sibling handlers):
  - `--disable-dev-shm-usage` / `--no-zygote` Chromium flags.
  - URL scheme + SSRF validation before *and* after navigation (defense in
    depth against private/internal network access).
  - A TCP preflight on the proxy (if any) so an unreachable/blocking proxy
    fails in ~8s instead of hanging for the full function timeout (same as
    ../serp-scraper).

Speed choices:
  - `humanize` defaults to `false` (no synthetic mouse/keyboard/scroll
    delay) — turn it on only for sites that need it.
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
from cloakbrowser import ProxySettings, launch
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

# How long to wait for a raw TCP handshake with the proxy before giving up.
# `cloakbrowser.launch(proxy=...)` has no timeout of its own — if the proxy
# silently drops the connection (common for residential-proxy providers that
# block/throttle datacenter-origin traffic, e.g. AWS Lambda's egress IPs) the
# whole invocation would otherwise hang until Lambda's function timeout,
# burning the full billed duration with no diagnostic signal. This preflight
# fails fast with a clear, actionable error instead. Same as ../serp-scraper.
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
    """Normalize a `proxy` event value into cloakbrowser/Playwright's
    `ProxySettings` shape: {"server": "<scheme>://<host>:<port>", "bypass"?,
    "username"?, "password"?}.

    This matters more than it looks: Chromium's underlying `--proxy-server`
    flag has no concept of credentials embedded in a URL. A caller-supplied
    string like "http://user:pass@host:port" connects to the proxy just fine
    but never authenticates — Chromium silently drops the embedded
    credentials, the proxy responds 407 Proxy Authentication Required, and
    the navigation hangs until Playwright/Lambda's own timeout. Splitting
    `username`/`password` into their own fields is what makes Playwright
    wire up its automatic proxy-auth handling. A dict already in that shape
    is passed through unchanged. Same logic as ../serp-scraper.
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
    """Raw TCP reachability check only — deliberately does not attempt proxy
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
            f"failure, not a credentials issue — it usually means the proxy "
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
        # Auto-derive timezone/locale from the proxy's exit IP so the browser
        # fingerprint matches where the traffic is actually coming from — a
        # mismatch against a residential proxy's geolocation is itself a
        # detection signal. Only meaningful when `proxy` is set.
        geoip=event.get("geoip", False),
        args=[
            # Lambda /dev/shm is ~64 MB — Chromium crashes mid-render without this.
            "--disable-dev-shm-usage",
            # Lambda's restricted process model can't fork from Chromium's
            # zygote — without this, child renderer processes fail to spawn.
            "--no-zygote",
        ],
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

        # Two independent uploads — run them concurrently instead of
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
