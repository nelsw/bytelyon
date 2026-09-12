"""AWS Lambda handler: Google search + SERP parsing by section.

Runs a single Google search and returns a JSON object with the results
grouped by section under `data`: sponsored results, sponsored products,
organic results, organic products, and similar (related) queries. The raw
page HTML and a full-page screenshot are always saved to S3 too.

Uses the **sync** cloakbrowser/Playwright API (`cloakbrowser.launch`), same
as ../news-scraper and ../aws_lambda_site_crawler.

IMPORTANT — read before using this against Google at any volume:

  1. Google aggressively IP-reputation-blocks automated traffic, independent
     of browser fingerprinting. Requests from datacenter IPs (including AWS
     Lambda's own IP ranges) are commonly served a "Our systems have
     detected unusual traffic from your computer network" interstitial
     instead of a results page — CloakBrowser's stealth Chromium does not
     get you past this on its own. Use the `proxy` event field with a
     quality residential/mobile proxy for reliable results. This handler
     detects the block page and logs a warning, but does not fail the
     invocation — the saved HTML/screenshot are the best way to confirm
     whether a given call was blocked, and `data` is returned all-empty
     rather than attempting to parse a block/CAPTCHA page.
  2. Google's SERP markup is unversioned, changes without notice, and uses
     largely obfuscated/rotating class names. The extraction in
     `_DATA_EXTRACT_JS` is deliberately **structural** (looks for long-lived
     attributes/tag names like `[data-pcu]` or `h3[id]` rather than hardcoded
     class names) so it degrades gracefully as markup shifts, but it is
     inherently best-effort. Expect to revisit it over time.
  3. Result/product links are resolved to their final destination URL with a
     real HTTP request per link (see `_final_url`), through the same browser
     context/proxy as the search itself. This adds latency proportional to
     the number of results on the page; a small in-memory cache
     (`_URL_CACHE`) avoids redundant lookups for repeated links.
  4. Scraping Google's SERPs may be subject to Google's Terms of Service.
     That's a legal/policy call for you to make for your use case — this
     example doesn't make it for you.

Event schema:
    query                             str        required — the search query
    hl                                str        "en"  — Google UI language
    gl                                str        "us"  — Google country
    num                               int        10    — requested result count (Google may ignore/cap this)
    proxy                             str|ProxySettings  none  — strongly recommended, see above.
                                                 Either "http://user:pass@host:port" (credentials
                                                 are split out automatically — see _normalize_proxy)
                                                 or a cloakbrowser/Playwright `ProxySettings` dict:
                                                 {"server": "socks5://host:port", "username": ...,
                                                  "password": ..., "bypass": ".google.com"}
    geoip                             bool       false — auto-derive timezone/locale from the
                                                 proxy's exit IP (only meaningful with `proxy` set)
    bucket                            str        "bytelyon-private" — S3 bucket for page
                                                 content + screenshot
    prefix                            str        "serp-scrapes/{aws_request_id}/" — S3 key
                                                 prefix for both uploads
    headless                         bool       true
    humanize                         bool       true
    human_preset                     str        "careful"
    goto_timeout_ms                  int        30000
    wait_for_load_state_timeout_ms   int         5000
    url_resolve_timeout_ms           int        10000 — per-link timeout when resolving
                                                 result/product links to their final destination

Returns:
    {
      "query": "...",
      "url": "https://www.google.com/search?q=...",
      "screenshot_key": "serp-scrapes/<request-id>/openai-<hash>.png",
      "content_key": "serp-scrapes/<request-id>/openai-<hash>.html",
      "data": {
        "sponsored_results": [
          {"kind": "sponsored_result", "index": 0, "title": "...", "brand": "...", "url": "...", "domain": "..."}
        ],
        "sponsored_products": [
          {"kind": "sponsored_product", "index": 0, "domain": "...", "url": "...", "image": "...", "title": "...", "price": "...", "brand": "..."}
        ],
        "organic_results": [
          {"kind": "organic_result", "index": 0, "title": "...", "url": "...", "domain": "..."}
        ],
        "organic_products": [
          {"kind": "organic_product", "index": 0, "title": "...", "image": "..."}
        ],
        "similar_queries": [
          {"kind": "similar_query", "index": 0, "value": "..."}
        ]
      }
    }

Every section in `data` degrades independently to `[]` on a parsing failure
(e.g. an unexpected markup change), and the whole `data` object is all-empty
when the page was blocked or extraction failed outright — see `_extract_data`
and `_empty_data`.

The full-page screenshot and the raw page HTML are always uploaded to S3
(same key stem, `.png` / `.html` extensions) regardless of whether Google
served real results or a block/CAPTCHA page — the latter is deliberately
still saved since it's the most useful artifact for diagnosing *why* a given
proxy/session got blocked. Both `screenshot_key` and `content_key` are
returned in the response.

Lambda-specific additions (see ../aws_lambda/INSTRUCTIONS.md for the same
rationale in the sibling handlers):
  - `--disable-dev-shm-usage` / `--no-zygote` Chromium flags.
  - `query` is URL-encoded and appended to a fixed `https://www.google.com/search`
    base — there's no SSRF surface here since the destination host isn't
    caller-controlled (unlike the other handlers, which navigate to
    caller-supplied URLs).
"""

from __future__ import annotations

import hashlib
import logging
import re
import socket
import time
from typing import Any
from urllib.parse import unquote, urlencode, urlparse

import boto3  # pyright: ignore[reportMissingImports]
from cloakbrowser import ProxySettings, launch  # pyright: ignore[reportMissingImports]
from playwright.sync_api import Page  # pyright: ignore[reportMissingImports]
from playwright.sync_api import (  # pyright: ignore[reportMissingImports]
    TimeoutError as PlaywrightTimeoutError,
)

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
# fails fast with a clear, actionable error instead.
_PROXY_PREFLIGHT_TIMEOUT_S = 8.0


def _normalize_proxy(proxy: Any) -> ProxySettings | None:
    """Normalize a `proxy` event value into cloakbrowser/Playwright's
    `ProxySettings` shape: {"server": "<scheme>://<host>:<port>", "bypass"?,
    "username"?, "password"?}.

    This matters more than it looks: Chromium's underlying `--proxy-server`
    flag has no concept of credentials embedded in a URL. A caller-supplied
    string like "http://user:pass@host:port" connects to the proxy just fine
    but never authenticates — Chromium silently drops the embedded
    credentials, the proxy responds 407 Proxy Authentication Required, and
    nothing in the stack retries with credentials or surfaces an error, so
    the navigation hangs until Playwright/Lambda's own timeout. Splitting
    `username`/`password` into their own fields is what makes Playwright
    wire up its automatic proxy-auth handling. A dict already in that shape
    (e.g. `{"server": "socks5://host:port", "username": ..., "password": ...,
    "bypass": ".google.com"}`) is passed through unchanged.
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

    server = f"{parsed.scheme}://{parsed.hostname}"
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


def _empty_data() -> dict:
    """All-empty `data` shape, used when the page is blocked or extraction
    fails outright, so callers can always rely on the same response shape.
    """
    return {
        "sponsored_results": [],
        "sponsored_products": [],
        "organic_results": [],
        "organic_products": [],
        "similar_queries": [],
    }


# Structural, best-effort SERP parser: looks for long-lived structural hooks
# (attributes/tag names) rather than Google's obfuscated/rotating class
# names, so it degrades gracefully as markup shifts. Runs in the page via
# `page.evaluate` and returns raw values only (hrefs, text) — resolving a
# link to its final destination requires a real HTTP request, which happens
# afterwards in `_final_url` since that can't happen synchronously from
# within `page.evaluate`.
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
# the same warm Lambda container. Google's organic/sponsored result links
# are frequently `/url?q=...`-style redirects rather than the destination
# itself; resolving one costs a real HTTP round-trip, and the same
# destination often repeats across a single SERP (or across nearby
# invocations against the same warm container), so caching meaningfully
# cuts down on redundant requests.
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
    """Best-effort structural SERP parser.

    Runs `_DATA_EXTRACT_JS` in the page to pull raw text/attributes, then
    resolves each result's link to its final destination URL. Each section
    is independently guarded — a failure in one (e.g. an unexpected markup
    change) falls back to an empty list for that section rather than
    failing the whole extraction.
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


def _is_blocked(page: Page) -> bool:
    if "/sorry/" in page.url:
        return True
    try:
        text = page.locator("body").inner_text(timeout=2_000).lower()
    except Exception:
        return False
    return any(phrase in text for phrase in _BLOCK_PHRASES)


def _s3_key(prefix: str, query: str, url: str, ext: str) -> str:
    safe_query = (
        re.sub(r"[^a-zA-Z0-9_-]+", "_", query.strip()).strip("_")[:80] or "query"
    )
    digest = hashlib.sha256(url.encode()).hexdigest()[:10]
    return f"{prefix.rstrip('/')}/{safe_query}-{digest}.{ext}"


def handler(event: dict, context: Any) -> dict:
    query = event.get("query")
    if not isinstance(query, str) or not query.strip():
        raise ValueError("event must include a non-empty 'query' string")

    params = {"q": query, "hl": event.get("hl", "en"), "gl": event.get("gl", "us")}
    if "num" in event:
        params["num"] = event["num"]
    search_url = "https://www.google.com/search?" + urlencode(params)

    bucket = event.get("bucket", DEFAULT_BUCKET)
    prefix = event.get(
        "prefix", f"serp-scrapes/{getattr(context, 'aws_request_id', 'local')}/"
    )
    s3 = boto3.client("s3")

    proxy: ProxySettings | None = _normalize_proxy(event.get("proxy"))
    _preflight_proxy(proxy)

    browser = launch(
        headless=event.get("headless", True),
        humanize=event.get("humanize", True),
        human_preset=event.get("human_preset", "careful"),
        proxy=proxy,
        # Auto-derive timezone/locale from the proxy's exit IP so the browser
        # fingerprint matches where the traffic is actually coming from —
        # a timezone/locale mismatch against a residential proxy's geolocation
        # is itself a detection signal. Only meaningful when `proxy` is set.
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
            page.goto(search_url, timeout=event.get("goto_timeout_ms", 30_000))
            page.wait_for_load_state(
                "domcontentloaded",
                timeout=event.get("wait_for_load_state_timeout_ms", 5_000),
            )
        except PlaywrightTimeoutError:
            pass

        _dismiss_consent(page)

        # Saved regardless of blocked/success outcome — useful for debugging
        # block pages too (the last several rounds of proxy debugging on this
        # handler would have been far faster with these on hand).
        html = page.content()
        png = page.screenshot(full_page=True)
        content_key = _s3_key(prefix, query, page.url, "html")
        screenshot_key = _s3_key(prefix, query, page.url, "png")
        s3.put_object(
            Bucket=bucket,
            Key=content_key,
            Body=html.encode("utf-8"),
            ContentType="text/html; charset=utf-8",
        )
        s3.put_object(
            Bucket=bucket, Key=screenshot_key, Body=png, ContentType="image/png"
        )

        if _is_blocked(page):
            logger.warning(
                "Google served a block/CAPTCHA page for query=%r — this is an "
                "IP-reputation block, not a fingerprinting issue; pass a "
                "residential/mobile proxy via the 'proxy' event field.",
                query,
            )
            data = _empty_data()
        else:
            data = _extract_data(
                page, url_resolve_timeout_ms=event.get("url_resolve_timeout_ms", 10_000)
            )

        return {
            "query": query,
            "screenshot_key": screenshot_key,
            "content_key": content_key,
            "data": data,
        }
    finally:
        browser.close()
