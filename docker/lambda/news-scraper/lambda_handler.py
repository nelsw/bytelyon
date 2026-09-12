"""AWS Lambda handler: batch news/article metadata scraping via cloakbrowser.

Adapted from a standalone CLI script that did:

    browser = launch(headless=True, humanize=True, human_preset="careful")
    page = browser.new_page()
    for url in sys.argv[1:]:
        page.goto(url)
        page.wait_for_load_state('domcontentloaded', timeout=5_000)
        data.append({...})
    print(json.dumps(data, indent=2))

The Lambda version takes `urls` (or `url`) from the event instead of
sys.argv, reuses a single browser/page across every URL in one invocation
(same as the source script), and returns the list of extracted records
directly instead of printing it.

Uses the **sync** cloakbrowser/Playwright API (`cloakbrowser.launch`), same
as the source script — this is fine under awslambdaric, which invokes
`handler(event, context)` synchronously with no asyncio event loop already
running. No Xvfb/headed display is required: the browser launches
headless=True by default, exactly like the source script.

Event schema:
    url                              str        a single URL to process
    urls                             list[str]   one or more URLs to process (checked before `url`)
                                                  Exactly one of `url` / `urls` is required.
    headless                         bool        True  (matches source script)
    humanize                         bool        True  (matches source script)
    human_preset                     str         "careful" (matches source script)
    goto_timeout_ms                  int         30000 — Playwright default; source script left this unset
    wait_for_load_state_timeout_ms   int         5000  (matches source script's hardcoded 5_000)

Returns:
    A JSON array, one object per URL (source script's `data` list):
        [{"url", "body", "img_src", "img_alt", "description", "keywords"}, ...]

    If navigation/extraction for a given URL raises an error other than a
    Playwright timeout (which the source script already tolerates), that
    URL's entry becomes {"url": ..., "error": "..."} instead of aborting the
    whole batch — added so one bad URL in a multi-URL request doesn't throw
    away results already gathered for the others.

Known quirks carried over from the source script (preserved for behavioral
fidelity, not fixed):
  - `body()`'s `for selector in [...]: locator = page.locator(selector); if
    locator is not None: break` always breaks on the first iteration —
    `page.locator(...)` returns a lazy handle and is never None, even when
    no element matches. In practice this means `body()` always resolves
    against the 'article' selector, never falling back to main/body/html.
  - The source script's `if txt is not "":` (identity comparison against a
    string literal) was changed to `if txt != "":` here since the former
    raises a SyntaxWarning on modern Python and is not reliable across
    interpreters; behavior is unchanged in practice.

Lambda-specific additions (not present in the source script, required for
correct operation in Lambda's environment — see ../aws_lambda/INSTRUCTIONS.md):
  - `--disable-dev-shm-usage` / `--no-zygote` Chromium flags.
  - URL scheme + SSRF validation before *and* after navigation (defense in
    depth against private/internal network access), matching the sibling
    ../aws_lambda handler.
"""

from __future__ import annotations

import ipaddress
import logging
import socket
from typing import Any
from urllib.parse import urlparse

from cloakbrowser import launch  # pyright: ignore[reportMissingImports]
from playwright.sync_api import Page  # pyright: ignore[reportMissingImports]
from playwright.sync_api import (  # pyright: ignore[reportMissingImports]
    TimeoutError as PlaywrightTimeoutError,
)

logger = logging.getLogger("cloakbrowser.lambda")
logger.setLevel(logging.INFO)

KEYWORD_ATTRS = ["property='article:tag'", "name='news_keywords'", "name='keywords'"]
DESCRIPTION_ATTRS = [
    "name='description'",
    "property='og:description'",
    "name='twitter:description'",
    "name='abstract'",
]
IMG_ALT_ATTRS = ["property='og:image:alt'", "name='twitter:image:alt'"]
IMG_SRC_ATTRS = [
    "name='image'",
    "property='og:image'",
    "property='og:image:secure_url'",
    "name='twitter:image'",
    "name='twitter:image:src'",
]


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


def meta_content(page: Page, attrs: list[str]) -> str:
    for attr in attrs:
        try:
            for txt in page.locator(f"meta[{attr}]").all():
                text = txt.get_attribute("content", timeout=1_500)
                if text is not None:
                    return text
        except PlaywrightTimeoutError:
            pass
    return ""


def body(page: Page) -> str:
    text = []
    for selector in ["article", "main", "body", "html"]:
        if page.locator(selector).count() == 0:
            continue
        for p in page.locator(selector).locator("p").all():
            txt = p.text_content()
            if txt is None:
                continue
            text.append(txt.strip())
    out = " ".join(text)
    out = out.replace(
        "This is a modal window. Beginning of dialog window. Escape will cancel and close the window. End of dialog window.",
        "",
    )
    return out


def keywords(page: Page) -> list[str]:
    words = meta_content(page, KEYWORD_ATTRS).split(",")
    for idx, word in enumerate(words):
        txt = word.strip()
        if txt != "":
            words[idx] = txt
    words.sort()
    return words


def _extract(page: Page, url: str) -> dict:
    return {
        "url": url,
        "body": body(page),
        "img_src": meta_content(page, IMG_SRC_ATTRS),
        "img_alt": meta_content(page, IMG_ALT_ATTRS),
        "description": meta_content(page, DESCRIPTION_ATTRS),
        "keywords": keywords(page),
    }


def _urls_from_event(event: dict) -> list[str]:
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


def _process_url(page: Page, url: str, event: dict) -> dict:
    # Validated per-URL (not upfront for the whole batch) so one bad/blocked
    # URL doesn't discard results already gathered for the others.
    _validate_url(url)

    try:
        page.goto(url, timeout=event.get("goto_timeout_ms", 30_000))
        page.wait_for_load_state(
            "domcontentloaded",
            timeout=event.get("wait_for_load_state_timeout_ms", 5_000),
        )
    except PlaywrightTimeoutError:
        # Matches the source script: a slow/never-settling page still gets a
        # best-effort extraction from whatever loaded so far.
        pass

    # Re-check after navigation/redirects — defense in depth against
    # server-side redirects to blocked destinations (not in the source
    # script; see module docstring).
    _validate_url(page.url)

    return _extract(page, url)


def handler(event: dict, context: Any) -> list[dict]:
    urls = _urls_from_event(event)

    logger.info("🏎️\turls=[%d]", len(urls))

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
    data: list[dict] = []
    try:
        for url in urls:
            try:
                data.append(_process_url(page, url, event))
                logger.info("🏆\t%s", url)
            except Exception as e:
                logger.warning("processing failed for %s: %s", url, e)
                data.append({"url": url, "error": str(e)})
    finally:
        browser.close()

    logger.info("🏁\turls=[%d/%d]", len(data), len(urls))

    return data
