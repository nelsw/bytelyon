"""Minimal, no-browser diagnostic Lambda: does a plain HTTP GET (via
`requests`, no Playwright/Chromium at all) directly or through a proxy, and
reports back timing + outcome for each URL tried.

This exists purely to answer one question: does authenticated proxy traffic
hang/fail specifically when it originates from AWS Lambda, independent of
whether the client is a full browser (cloakbrowser/Chromium) or a plain
Python HTTP client? See ../serp-scraper/INSTRUCTIONS.md and this repo's
conversation history for the browser-based (Chromium) and httpx-based
failures this is meant to cross-check against.

Event schema:
    urls    list[str]              default: a small fixed set of IP-echo
                                    services + a Google search URL
    proxy   str | None             "http://user:pass@host:port" (or
                                    "socks5://user:pass@host:port"), applied
                                    to both http:// and https:// requests.
                                    None/omitted = direct, no proxy.
    timeout_s   float              default 25 \u2014 per-request timeout

Returns:
    {"proxy": "<redacted-or-null>", "results": [
        {"url": ..., "ok": bool, "status_code": int|null, "body": str|null,
         "error": str|null, "elapsed_s": float}
    ]}
"""

from __future__ import annotations

import time
from typing import Any
from urllib.parse import urlparse

import requests  # pyright: ignore[reportMissingImports]

DEFAULT_URLS = [
    "https://api.ipify.org",
    "https://checkip.amazonaws.com",
    "https://www.google.com/search?q=sailing+blocks",
]


def _redact_proxy(proxy: str | None) -> str | None:
    if not proxy:
        return None
    parsed = urlparse(proxy)
    if not parsed.hostname:
        return proxy
    netloc = parsed.hostname
    if parsed.port:
        netloc += f":{parsed.port}"
    return f"{parsed.scheme}://{netloc}"


def handler(event: dict, context: Any) -> dict:
    urls = event.get("urls") or DEFAULT_URLS
    proxy = event.get("proxy")
    timeout_s = float(event.get("timeout_s", 25))

    proxies = {"http": proxy, "https": proxy} if proxy else None

    results = []
    for url in urls:
        started = time.monotonic()
        entry: dict[str, Any] = {"url": url}
        try:
            resp = requests.get(url, proxies=proxies, timeout=timeout_s)
            entry["ok"] = True
            entry["status_code"] = resp.status_code
            entry["body"] = resp.text[:500]
            entry["error"] = None
        except Exception as e:  # noqa: BLE001 - report every failure mode
            entry["ok"] = False
            entry["status_code"] = None
            entry["body"] = None
            entry["error"] = f"{type(e).__name__}: {e}"
        entry["elapsed_s"] = round(time.monotonic() - started, 2)
        results.append(entry)

    return {"proxy": _redact_proxy(proxy), "results": results}
