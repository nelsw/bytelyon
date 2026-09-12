"""AWS Lambda handler: single-page scrape with full-page screenshot to S3.

Given a single URL, this handler:
  1. Navigates to the page.
  2. Takes a full-page screenshot and uploads it to S3.
  3. Returns the page's links, meta tags, final URL, and title as JSON.

Uses the **sync** cloakbrowser/Playwright API (`cloakbrowser.launch`), same
as ../news-scraper, ../aws_lambda_site_crawler, and
../aws_lambda_serp_scraper.

Event schema:
    url                              str        required — the page to scrape (http/https only)
    bucket                           str        "bytelyon-private"
    prefix                           str        "page-scrapes/{aws_request_id}/" — S3 key prefix
    headless                         bool       true
    humanize                         bool       true
    human_preset                    str        "careful"
    goto_timeout_ms                  int        30000
    wait_for_load_state_timeout_ms   int         5000
    include_links                    bool       true — when false, link extraction is
                                                 skipped entirely and "links" is omitted
                                                 from the response

Returns:
    {
      "url": "https://example.com/",
      "title": "Example Domain",
      "links": ["https://example.com/about", ...],
      "meta": [{"name": "description", "content": "..."}, ...],
      "screenshot_s3_uri": "s3://bytelyon-private/page-scrapes/.../example_com-<hash>.png"
    }

`links` only includes same-domain links (the linked host must match the
page's own host, ignoring a `www.` prefix on either side) and excludes any
link with a URL fragment (`#...`). Among the remaining links, `www.` vs.
non-`www.` variants of the same URL are deduplicated (e.g.
"https://www.example.com/x" and "https://example.com/x" are treated as the
same link) — the first-encountered form (in DOM order) is kept. Pass
`"include_links": false` in the event to skip link extraction altogether
(no DOM walk over anchors) and omit `links` from the response. `meta`
includes every `<meta>` tag on the page (whichever of `name`/`property`/
`http-equiv`/`charset` is present, plus `content`), not a curated subset —
see ../news-scraper for a handler that extracts a curated
set of OpenGraph/Twitter/news fields instead.

`screenshot_s3_uri` is not part of the literal ask but is included since
otherwise there'd be no way to locate the uploaded screenshot from the
response alone.

Lambda-specific additions (see ../aws_lambda/INSTRUCTIONS.md for the same
rationale in the sibling handlers):
  - `--disable-dev-shm-usage` / `--no-zygote` Chromium flags.
  - URL scheme + SSRF validation before *and* after navigation (defense in
    depth against private/internal network access).
"""

from __future__ import annotations

import hashlib
import ipaddress
import logging
import re
import socket
from typing import Any
from urllib.parse import urlparse

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

# Structural extraction: same-domain, fragment-free page links (www./non-www.
# deduplicated) + every <meta> tag on the page. Link extraction is skipped
# entirely (no DOM walk over anchors) when `includeLinks` is falsy.
_PAGE_EXTRACT_JS = r"""
(opts) => {
  const includeLinks = !!(opts && opts.includeLinks);
  const links = [];

  if (includeLinks) {
    const seen = new Set();
    const pageHost = location.host.replace(/^www\./i, "");

    document.querySelectorAll("a[href]").forEach((a) => {
      const href = a.href;
      if (!href) return;

      let u;
      try {
        u = new URL(href);
      } catch (e) {
        return; // a.href is normally already absolute; skip if unparseable.
      }

      // Check the raw href for a literal '#' rather than u.hash: an href of
      // "...#" (empty fragment) parses to u.hash === "" (falsy) but still
      // contains a fragment delimiter, which should be omitted too. a.href is
      // the browser's fully-resolved URL, so any literal '#' in a path segment
      // would already be percent-encoded as %23 -- a raw '#' here always means
      // an actual fragment delimiter.
      if (href.includes("#")) return; // omit links with fragments

      const host = u.host.replace(/^www\./i, "");
      if (host !== pageHost) return; // omit cross-domain links

      const key = u.protocol + "//" + host + u.pathname + u.search;
      if (!seen.has(key)) {
        seen.add(key);
        links.push(href);
      }
    });
  }

const meta = {};
Array.from(document.querySelectorAll('meta'))
    .forEach((m) => {
        const v = m.getAttribute('content')?.trim();
        const n = m.getAttribute('name')?.trim();
        const p = m.getAttribute('property')?.trim();
        if (v && (n || p)) {
            meta[n ? n : p] = v
        }
    });
console.log(meta);

  return { links, meta };
}
"""


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


def _domain(url: str) -> str:
    """Return the registrable-ish domain for display: hostname minus a leading
    'www.' (e.g. "https://www.publix.com/takeout?query=foo" -> "publix.com").
    """
    hostname = urlparse(url).hostname or ""
    return re.sub(r"^www\.", "", hostname, flags=re.IGNORECASE)


def _s3_key(prefix: str, url: str) -> str:
    parsed = urlparse(url)
    path = parsed.path.strip("/") or "index"
    safe = re.sub(r"[^a-zA-Z0-9/_-]", "_", f"{parsed.netloc}/{path}")
    digest = hashlib.sha256(url.encode()).hexdigest()[:10]
    return f"{prefix.rstrip('/')}/{safe}-{digest}.png"


def handler(event: dict, context: Any) -> dict:
    url = event.get("url")
    if not isinstance(url, str) or not url.strip():
        raise ValueError("event must include a non-empty 'url' string")
    _validate_url(url)

    bucket = event.get("bucket", DEFAULT_BUCKET)
    prefix = event.get(
        "prefix", f"page-scrapes/{getattr(context, 'aws_request_id', 'local')}/"
    )

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
    try:
        try:
            page.goto(url, timeout=event.get("goto_timeout_ms", 30_000))
            page.wait_for_load_state(
                "domcontentloaded",
                timeout=event.get("wait_for_load_state_timeout_ms", 5_000),
            )
        except PlaywrightTimeoutError:
            # Best-effort: screenshot/extract from whatever loaded so far.
            pass

        _validate_url(page.url)  # re-check after redirects

        png = page.screenshot(full_page=True)
        key = _s3_key(prefix, page.url)
        s3.put_object(Bucket=bucket, Key=key, Body=png, ContentType="image/png")

        include_links = event.get("include_links", True)
        extracted = page.evaluate(_PAGE_EXTRACT_JS, {"includeLinks": include_links})

        result = {
            "url": page.url,
            "domain": _domain(page.url),
            "title": page.title(),
            "meta": extracted["meta"],
            "screenshot_key": key,
        }
        if include_links:
            result["links"] = extracted["links"]
        return result
    finally:
        browser.close()
