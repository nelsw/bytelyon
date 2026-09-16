"""Shared helpers for every job-type handler under docker/worker/handlers/.

Deliberately minimal: this worker fleet's whole job is "go get this page
(optionally via a proxy/search box), save the HTML and a screenshot to S3,
report back {url, screenshot_key, content_key}." No parsing, no business
logic, no per-domain knowledge lives here or in any handler — that all moved
to the main Laravel app (see app/Support/Html.php, app/Data/Html/Page.php,
app/Support/Serp.php), so every worker's response payload is identical
regardless of job type, and the parsing logic lives in one place, in one
language, versioned with the rest of the app instead of scattered across
Python handlers that are harder to test/iterate on than the PHP app itself.
"""

from __future__ import annotations

import hashlib
import re
import uuid
from urllib.parse import urlparse

import boto3

DEFAULT_BUCKET = "bytelyon-private"


def proxy_settings(job: dict) -> dict | None:
    """Convert the bot's selected-proxy payload -- added to the SQS message
    by `App\\Jobs\\BotJob::proxyPayload()` on the Laravel side as
    `{scheme, host, port, username, pass, bypass}` -- into the
    Playwright-compatible `ProxySettings` shape (`{server, username?,
    password?, bypass?}`) that both `cloakbrowser.launch()` and
    `seleniumbase_playwright.launch()` accept as their `proxy` kwarg.

    `BotJob` picks one proxy at random (server-side, in PHP) whenever a bot
    has more than one configured -- this worker only ever sees the single
    already-chosen proxy for a given job, if any. Returns None when the job
    carries no `proxy` field at all (the bot has none configured), so
    callers can pass the result straight through as `proxy=...` without a
    None-check of their own.
    """
    proxy = job.get("proxy")
    if not proxy or not proxy.get("host"):
        return None

    scheme = proxy.get("scheme") or "http"
    port = proxy.get("port")
    server = f"{scheme}://{proxy['host']}" + (f":{port}" if port else "")

    settings: dict = {"server": server}
    if proxy.get("username"):
        settings["username"] = proxy["username"]
    if proxy.get("pass"):
        settings["password"] = proxy["pass"]
    if proxy.get("bypass"):
        settings["bypass"] = proxy["bypass"]
    return settings


def s3_key(prefix: str, url: str, ext: str) -> str:
    parsed = urlparse(url)
    path = parsed.path.strip("/") or "index"
    safe = re.sub(r"[^a-zA-Z0-9/_-]", "_", f"{parsed.netloc}/{path}")
    digest = hashlib.sha256(url.encode()).hexdigest()[:10]
    return f"{prefix.rstrip('/')}/{safe}-{digest}.{ext}"


def default_prefix(job_type: str) -> str:
    return f"worker-scrapes/{job_type}/{uuid.uuid4()}/"


def capture_and_upload(page, job_type: str, bucket: str = DEFAULT_BUCKET) -> dict:
    """Capture the current page's HTML + full-page screenshot and upload
    both to S3. Returns the generic `{url, screenshot_key, content_key}`
    contract every handler returns — the *only* thing worker.py's callback
    POST needs, regardless of job type.
    """
    url = page.url
    html = page.content()
    png = page.screenshot(full_page=True)

    prefix = default_prefix(job_type)
    content_key = s3_key(prefix, url, "html")
    screenshot_key = s3_key(prefix, url, "png")

    s3 = boto3.client("s3")
    s3.put_object(
        Bucket=bucket,
        Key=content_key,
        Body=html.encode("utf-8"),
        ContentType="text/html; charset=utf-8",
    )
    s3.put_object(
        Bucket=bucket,
        Key=screenshot_key,
        Body=png,
        ContentType="image/png",
    )

    return {"url": url, "screenshot_key": screenshot_key, "content_key": content_key}
