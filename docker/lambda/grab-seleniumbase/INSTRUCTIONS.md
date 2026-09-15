# SeleniumBase fast single-page capture on AWS Lambda

The [SeleniumBase](https://seleniumbase.io/) counterpart of
[`../grab`](../grab): given a single URL (with an optional proxy), uploads
the rendered HTML and a full-page screenshot to S3, and returns just the two
S3 keys. Same event schema, same S3 upload behavior, same response shape as
`../grab` -- the only difference is the stealth-browser provider.

This handler builds `FROM` [`../../images/seleniumbase-base`](../../images/seleniumbase-base)
instead of `../base-image`. That base image bakes in system Chromium,
`seleniumbase`, `playwright` (client only), and a `seleniumbase_playwright.launch()`
shim that launches SeleniumBase's stealthy CDP Mode Chromium and hands back
a normal `playwright.sync_api.Browser`, connected to it over the Chrome
DevTools Protocol ("Stealthy Playwright Mode" -- see the
[SeleniumBase README](https://seleniumbase.io/)). Every line of this
handler past the `launch()` call is otherwise identical to `../grab`,
because both providers hand back a standard Playwright `Page`.

## Files in this directory

| File                | Purpose                                                                                                 |
| ------------------- | ------------------------------------------------------------------------------------------------------- |
| `Dockerfile`        | `FROM` the shared `../../images/seleniumbase-base`, plus a single `COPY lambda_handler.py`.             |
| `lambda_handler.py` | Handler. Takes `{url, proxy?, ...}`, returns `{url, content_key, screenshot_key}`. Headless by default. |
| `INSTRUCTIONS.md`   | This file.                                                                                              |

This directory depends on the shared base image
([`../../images/seleniumbase-base`](../../images/seleniumbase-base)) having
been built and either published to ECR or loaded locally first -- see that
directory's `INSTRUCTIONS.md`.

## Build

Build and publish `../../images/seleniumbase-base` first (see its
`INSTRUCTIONS.md`), then:

```bash
docker buildx build --platform linux/arm64 --provenance=false --sbom=false \
  -t bytelyon-grab-seleniumbase:arm64 --load .
```

`--provenance=false --sbom=false` is required -- without it, buildx produces
a manifest-list/attestation image that Lambda's `CreateFunction` rejects
with `InvalidParameterValueException: ... image manifest, config or layer
media type ... is not supported`.

To build against a locally-loaded (not-yet-pushed) base image instead of
pulling from ECR:

```bash
docker buildx build --platform linux/arm64 --provenance=false --sbom=false \
  --build-arg BASE_IMAGE=bytelyon-seleniumbase-base:arm64 \
  -t bytelyon-grab-seleniumbase:arm64 --load .
```

For x86_64, switch `--platform linux/amd64` (and build the base image for
the same architecture). Unlike CloakBrowser's official image, the
SeleniumBase base image's system Chromium supports both architectures
equally, since it comes from Debian's `chromium` package rather than Google
Chrome (amd64-only).

## Local smoke test (no Lambda deploy needed, but real AWS credentials required)

Both artifacts are genuinely uploaded to S3 even in local RIE testing, so
pass through real AWS credentials with write access to the target bucket:

```bash
docker run --rm -p 9000:8080 \
  -e AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY -e AWS_SESSION_TOKEN -e AWS_REGION=us-east-1 \
  bytelyon-grab-seleniumbase:arm64

# In another shell:
curl -sS -XPOST "http://localhost:9000/2015-03-31/functions/function/invocations" \
  -d '{"url":"https://example.com"}'

# With a proxy:
curl -sS -XPOST "http://localhost:9000/2015-03-31/functions/function/invocations" \
  -d '{"url":"https://example.com", "proxy":"http://user:pass@host:port"}'
```

## Event schema

| Field                  | Type      | Default                                                                                                     |
| ---------------------- | --------- | ----------------------------------------------------------------------------------------------------------- |
| `url`                  | str       | required -- the page to capture (http/https only)                                                           |
| `proxy`                | str\|dict | none -- `"http://user:pass@host:port"` (credentials split out automatically) or a `ProxySettings` dict      |
| `geoip`                | bool      | `false` -- accepted for parity with `../grab`; **no-op** with this provider (see below)                     |
| `bucket`               | str       | `"bytelyon-private"`                                                                                        |
| `prefix`               | str       | `"grabs/{aws_request_id}/"` -- S3 key prefix for both uploads                                               |
| `headless`             | bool      | `true`                                                                                                      |
| `humanize`             | bool      | `false` -- accepted for parity with `../grab`; **no-op** with this provider (see below)                     |
| `human_preset`         | str       | `"careful"` -- accepted for parity; **no-op**                                                               |
| `wait_until`           | str       | `"domcontentloaded"` -- passed to `page.goto` (`"load"`\|`"domcontentloaded"`\|`"networkidle"`\|`"commit"`) |
| `goto_timeout_ms`      | int       | `30000`                                                                                                     |
| `full_page_screenshot` | bool      | `true`                                                                                                      |

### Response

```json
{
    "url": "https://example.com/",
    "content_key": "grabs/<request-id>/example_com-<hash>.html",
    "screenshot_key": "grabs/<request-id>/example_com-<hash>.png"
}
```

## Differences from `../grab` (CloakBrowser)

- **Stealth mechanism**: SeleniumBase's CDP Mode launch flags/patches
  instead of CloakBrowser's own stealth Chromium build. Neither this
  handler nor `../grab` makes any claim about which is more effective
  against any particular anti-bot vendor -- pick based on what works for
  your target sites.
- **`humanize` / `human_preset` are no-ops** -- these are CloakBrowser
  concepts (synthetic mouse/keyboard/scroll timing). Accepted here purely
  so callers can point either handler at the same event payload.
- **`geoip` is a no-op** -- CloakBrowser auto-derives timezone/locale from
  the proxy's exit IP; SeleniumBase has no built-in equivalent. If you need
  specific timezone/locale values, this handler would need extending to
  accept and forward explicit `timezone`/`locale` fields (both are already
  accepted by `seleniumbase_playwright.launch(**kwargs)` -- see
  `../../images/seleniumbase-base/seleniumbase_playwright.py`).
- **No chromedriver** -- SeleniumBase's CDP Mode talks directly to Chromium
  over the DevTools Protocol, so the base image doesn't install one.

## Lambda-specific additions

- **Chromium hardening (do not remove)**: `--no-zygote` and `--disable-gpu`
  are always applied by `seleniumbase_playwright.launch()`
  (`--disable-dev-shm-usage` is already one of SeleniumBase's own default
  CDP Mode args, and `--no-sandbox` is always forced on for the same
  reason -- see that module's docstring). Notably, `--single-process` is
  deliberately **not** applied, despite being a common recommendation for
  "Chromium on Lambda" setups: confirmed by testing against a real deployed
  Lambda function, it actively breaks launches here (Chromium starts but
  its CDP debug port never opens) -- see the long comment above the flag
  list in `seleniumbase_playwright.py` for the full story. This only
  reproduced in real Lambda, never in a plain `docker run` of the same
  image, so if you're debugging a _different_ launch problem, don't trust
  a local `docker run`/RIE pass alone -- test against a real deployed
  function too.
- **Launch retries + captured Chromium output**: `seleniumbase_playwright.launch()`
  retries the actual Chromium launch up to 3 times (a fresh Lambda
  execution environment's very first launch attempt is occasionally flaky
  even with correct arguments -- observed empirically, not just
  theorized), and on final failure raises with whatever the Chromium
  subprocess itself printed to stdout/stderr (normally discarded to
  DEVNULL by SeleniumBase), turning a generic "Failed to connect to the
  browser" into an actionable error.
- **URL validation**: scheme + SSRF checks on the input `url` before
  navigation, and again on `page.url` after navigation/redirects.
- **Proxy TCP preflight**: if `proxy` is set, a raw TCP connect to the
  proxy's host:port is attempted first (~8s timeout) so an
  unreachable/blocking proxy fails fast with a clear error instead of
  hanging until the function timeout. Same as `../grab` and
  `../serp-scraper`.

## IAM permissions

In addition to `AWSLambdaBasicExecutionRole` (CloudWatch Logs), the
execution role needs write access to the target bucket:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": ["s3:PutObject"],
            "Resource": ["arn:aws:s3:::bytelyon-private/*"]
        }
    ]
}
```

## Function configuration recommendations

| Setting                    | Value                                                             | Why                                                                                                           |
| -------------------------- | ----------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------- |
| Package type               | Image                                                             | Required -- this is a container image, not a zip.                                                             |
| Architecture               | `arm64`                                                           | Roughly 20% cheaper than x86_64. Match the architecture you built for.                                        |
| Memory                     | 3008 MB                                                           | Memory in Lambda is tied to vCPU. Below ~1769 MB Chromium starts noticeably slower.                           |
| Timeout                    | 30-45 s                                                           | Single page load + full-page screenshot + concurrent uploads is fast once warm; budget for cold start on top. |
| Ephemeral storage (`/tmp`) | 512 MB (default)                                                  | Both artifacts are held in memory and streamed to S3, not accumulated on disk.                                |
| Networking                 | Default (no VPC)                                                  | Only outbound HTTPS to the target site (or proxy) and to S3 is needed.                                        |
| Execution role             | Basic execution + S3 `PutObject` on the target bucket (see above) |                                                                                                               |

## Cold start

Measured against a real deployed Lambda function (`arm64`, 3008 MB): first
invocation in a new execution environment typically completes in
**~4-10s** end-to-end (Lambda's own `Init Duration` for this image is
~2.3s; the rest is the Chromium launch + page load + S3 upload). Warm
invocations (same execution environment, fresh Chromium process each time
-- this handler doesn't persist a browser across invocations) run in
**~3-4s**. This is notably faster than `../grab`'s documented ~80-90s
cold start -- unclear whether that's inherent to CloakBrowser's own bundled
Chromium vs. this image's Debian `chromium` package, image size/layer
count differences, or something else; take both numbers as rough guidance
rather than a rigorous head-to-head benchmark.

Occasionally (empirically, roughly 1 in 4-5 cold starts in initial
testing) a fresh execution environment's _very first_ Chromium launch
attempt fails outright before the debug port ever opens, while a retry in
the same, now-warm container succeeds immediately -- `seleniumbase_playwright.launch()`
retries internally for exactly this reason (see "Lambda-specific
additions" above), so this shouldn't surface as a handler-level failure,
but it does mean an occasional cold start takes a few seconds longer than
the happy path above while a retry plays out.
