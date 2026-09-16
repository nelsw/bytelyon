# CloakBrowser Google SERP scraper on AWS Lambda

Runs a Google search and parses the results page into a JSON object grouped by SERP section (organic results, featured snippet, People Also Ask, related searches, ads, knowledge panel), using CloakBrowser's stealth Chromium inside an AWS Lambda function (container image package type).

This is a sibling of [`../news-scraper`](../news-scraper), [`../page-scraper`](../page-scraper), [`../scraper`](../scraper), and [`../site-crawler`](../site-crawler) — all five build `FROM` the shared [`../base-image`](../base-image), which carries the Dockerfile/entrypoint scaffolding derived from the official `cloakhq/cloakbrowser` Docker Hub image that used to be duplicated in every handler's own `Dockerfile`. Every other invocation surface from the canonical image (`python`, `cloakserve`, `cloaktest`, `node`, `bash`, examples) keeps working.

## Read this before using it against Google at any volume

1. **Google IP-reputation-blocks automated traffic, independent of browser fingerprinting.** Requests from datacenter IPs — including AWS Lambda's own IP ranges — are commonly served a "Our systems have detected unusual traffic from your computer network" interstitial instead of a results page. This was confirmed while building this example: identical requests from a shared/cloud egress IP were blocked outright, with no captcha to solve — a hard IP-based block. CloakBrowser's stealth Chromium (fingerprint patches, humanize) does not get you past this on its own; it addresses bot _detection_, not IP _reputation_. Use the `proxy` event field with a quality residential/mobile proxy for reliable results.
2. Because of (1), this handler was built and tested **structurally** rather than against a live, verified-working SERP capture — see "How the parser works" below. Treat the section-extraction logic as best-effort, and expect to revisit it as you observe real results (or as Google's markup shifts).
3. Google's SERP markup is unversioned, changes without notice, and leans on obfuscated/rotating class names. There is no long-term-stable public API for this — that's exactly why the parser favors structural patterns (e.g. "an `<h3>` near a link" = an organic result) over hardcoded class names, but it's inherently fragile compared to a real API.
4. **Scraping Google's SERPs may be subject to Google's Terms of Service.** That's a legal/policy call for your use case — this example doesn't make it for you.

## A note on proxy credentials

Chromium's underlying `--proxy-server` flag has no concept of credentials embedded in a URL. If you pass `"proxy": "http://user:pass@host:port"`, the handler automatically splits the credentials into separate `username`/`password` fields before launching (see `_normalize_proxy` in `lambda_handler.py`, which returns a `cloakbrowser.ProxySettings` dict) — this is what actually makes proxy authentication work. Without that split, Chromium connects to the proxy, silently drops the embedded credentials, gets a `407 Proxy Authentication Required` response, and the navigation hangs with no clear error until the Lambda function times out. If you already have a `ProxySettings`-shaped dict (`{"server", "username", "password", "bypass"}`), pass it as-is — it's used unchanged. Dict form is also the only way to set `bypass` or use a non-`http` proxy scheme (e.g. `socks5://`).

**Don't set `bypass` to `.google.com` (or anything matching it).** `bypass` tells Chromium which hosts should skip the proxy and connect _directly_ instead — the opposite of what you want here, since the whole point of `proxy` in this handler is to route the Google search request through it. Bypassing google.com sends that request straight from Lambda's own datacenter IP, which is exactly the IP-reputation block described above. Only use `bypass` for other hosts you specifically want excluded from the proxy.

A fast TCP preflight (`_preflight_proxy`, ~8s timeout) runs before every browser launch and raises a clear error if nothing is listening on the proxy's host:port, instead of silently hanging for the full function timeout. It only checks that a TCP connection can be opened — it does not attempt authentication or a CONNECT tunnel, so a pass here doesn't guarantee the proxy will actually route traffic.

## Retries on a bad proxy connection

Rotating residential/mobile proxies typically hand out a fresh exit IP per TCP connection, with no session stickiness. At any given moment some fraction of exit peers in a shared pool are already rate-limited or blocked by Google — this was confirmed empirically: a small back-to-back sample of plain HTTP requests through a real residential proxy came back 3x `200`, 1x `429`, 1x `403` straight from Google, with the proxy connection itself succeeding every single time in under 1.5s (no proxy-side blocking or slowness at all). A full browser session opens far more concurrent connections than a single request, so it's proportionally more likely to land at least one bad peer — usually surfacing as a hard `page.goto` failure (`net::ERR_CONNECTION_CLOSED`, `net::ERR_TIMED_OUT`, etc.) rather than a slow load.

On a hard navigation failure, the handler closes the browser and relaunches a fresh one — a new browser means a new proxy connection, which means a new (hopefully clean) exit IP — up to `goto_retry_attempts` (default `3`) total attempts before giving up and raising. A `PlaywrightTimeoutError` (page loaded, just slowly) does **not** count as a failure and is not retried; only a hard connection-level error does. A CAPTCHA/block page is also retried (not just hard connection failures) for the same reason — see `_is_blocked` usage in `lambda_handler.py`.

## Oxylabs: skip country targeting

`App\Services\LambdaService::proxyUsername()` normalizes the `customer-` username prefix Oxylabs' targeting suffixes require, and `lambda_handler.py`'s `_oxylabs_sticky_proxy()` appends a fresh `-sessid-<id>` per retry attempt (see its docstring for why that has to happen per-attempt, not once per invocation). Deliberately **not** included: a `-cc-<country>` suffix. It's tempting — tighter `geoip` timezone/locale matching — but confirmed empirically to make things measurably worse: several consecutive Google blocks in a row with `-cc-US` set, across different exit IPs and different queries (ruling out a single bad IP or query-specific rate-limiting), immediately followed by a clean success the moment `-cc-US` was removed and the plain username used instead. Whatever sub-pool Oxylabs routes `-cc-US` traffic through, it was smaller/more commonly flagged than the default pool at the time this was tested. Re-verify carefully (small sample sizes here, and proxy pool quality can drift over time) before reintroducing country targeting.

## Files in this directory

| File                | Purpose                                                                                                                                          |
| ------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------ |
| `Dockerfile`        | `FROM` the shared `../base-image`, plus a single `COPY lambda_handler.py`. See `../base-image/INSTRUCTIONS.md` for what the base image provides. |
| `lambda_handler.py` | Search + parse handler. Takes `{query, ...}`, returns a JSON object keyed by SERP section. Headed by default (see below).                        |
| `INSTRUCTIONS.md`   | This file.                                                                                                                                       |

This directory is no longer fully standalone: it depends on the shared base image (`../base-image`) having been built and either published to ECR or loaded locally first — see `../base-image/INSTRUCTIONS.md`. Everything specific to this handler (its Dockerfile and `lambda_handler.py`) still lives entirely in this directory.

## Build

Build and publish `../base-image` first (see `../base-image/INSTRUCTIONS.md`), then:

```bash
docker buildx build --platform linux/arm64 --provenance=false --sbom=false \
  -t bytelyon-serp-scraper:arm64 --load .
```

`--provenance=false --sbom=false` is required — without it, buildx produces a manifest-list/attestation image that Lambda's `CreateFunction` rejects with `InvalidParameterValueException: ... image manifest, config or layer media type ... is not supported`.

To build against a locally-loaded (not-yet-pushed) base image instead of pulling from ECR, add `--build-arg BASE_IMAGE=bytelyon-cloakbrowser-base:arm64`. For x86_64, switch `--platform linux/amd64` (and build the base image for the same architecture).

## Local smoke test (no AWS account needed)

```bash
docker run --rm -p 9000:8080 bytelyon-serp-scraper:arm64

# In another shell:
curl -sS -XPOST "http://localhost:9000/2015-03-31/functions/function/invocations" \
  -d '{"query":"openai"}'

# With a proxy (recommended — see above), string form:
curl -sS -XPOST "http://localhost:9000/2015-03-31/functions/function/invocations" \
  -d '{"query":"openai","proxy":"http://user:pass@proxy-host:port"}'

# ...or dict form (required for non-http proxy schemes, e.g. socks5, and to
# set `bypass`; also what the string form above is normalized into internally).
# Do NOT set `bypass` to ".google.com" here — see "A note on proxy credentials" above:
curl -sS -XPOST "http://localhost:9000/2015-03-31/functions/function/invocations" \
  -d '{"query":"openai","proxy":{"server":"socks5://proxy-host:port","username":"user","password":"pass"}}'
```

Other invocation surfaces stay intact (these match the canonical CloakHQ image):

```bash
docker run --rm -it bytelyon-serp-scraper:arm64 python                          # REPL
docker run --rm bytelyon-serp-scraper:arm64 python examples/basic.py            # examples
docker run --rm -p 9222:9222 bytelyon-serp-scraper:arm64 cloakserve --port=9222 # CDP server
docker run --rm bytelyon-serp-scraper:arm64 cloaktest                           # stealth tests
docker run --rm -it bytelyon-serp-scraper:arm64 node                            # JS wrapper
```

## Event schema

| Field                            | Type                  | Default                                                                                                                                                                                                                                                                                                                                                                                                            |
| -------------------------------- | --------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `query`                          | str                   | required — the search query                                                                                                                                                                                                                                                                                                                                                                                        |
| `hl`                             | str                   | `"en"` — Google UI language                                                                                                                                                                                                                                                                                                                                                                                        |
| `gl`                             | str                   | `"us"` — Google country                                                                                                                                                                                                                                                                                                                                                                                            |
| `num`                            | int                   | `10` — requested result count (Google may ignore or cap this)                                                                                                                                                                                                                                                                                                                                                      |
| `proxy`                          | str / `ProxySettings` | none — **strongly recommended**, see above. Either `"http://user:pass@host:port"` (credentials are split out into `username`/`password` automatically — see "A note on proxy credentials" below) or a `cloakbrowser.ProxySettings` dict: `{"server": "socks5://host:port", "username": ..., "password": ...}`. Do **not** set `bypass` to `.google.com` — see "A note on proxy credentials" above.                 |
| `geoip`                          | bool                  | `false` — auto-derive timezone/locale from the proxy's exit IP; only meaningful when `proxy` is set. A timezone/locale that doesn't match the proxy's geolocation is itself a detection signal.                                                                                                                                                                                                                    |
| `headless`                       | bool                  | `false` — headed via Xvfb, baked into the base image (`:99`, `DISPLAY=:99`). CloakBrowser's own guidance: "most blocks come from missing" a residential proxy, `geoip`, or headed mode — "some sites detect headless even with our C++ patches."                                                                                                                                                                   |
| `humanize`                       | bool                  | `true`                                                                                                                                                                                                                                                                                                                                                                                                             |
| `human_preset`                   | str                   | `"careful"`                                                                                                                                                                                                                                                                                                                                                                                                        |
| `goto_timeout_ms`                | int                   | `30000`                                                                                                                                                                                                                                                                                                                                                                                                            |
| `goto_retry_attempts`            | int                   | `3` — rotating residential/mobile proxies hand out a fresh exit IP per connection with no stickiness, and any individual peer may already be rate-limited/blocked by Google independent of this proxy/account (confirmed empirically — see below). On a hard navigation failure the browser is closed and relaunched — a fresh proxy connection means a fresh exit IP — up to this many attempts before giving up. |
| `wait_for_load_state_timeout_ms` | int                   | `5000`                                                                                                                                                                                                                                                                                                                                                                                                             |
| `url_resolve_timeout_ms`         | int                   | `10000` — per-link timeout when resolving a result/product link to its final destination URL                                                                                                                                                                                                                                                                                                                       |
| `bucket`                         | str                   | `"bytelyon-private"` — S3 bucket for the saved page content + screenshot                                                                                                                                                                                                                                                                                                                                           |
| `prefix`                         | str                   | `"serp-scrapes/{aws_request_id}/"` — S3 key prefix for both uploads                                                                                                                                                                                                                                                                                                                                                |

### Response

```json
{
    "query": "openai",
    "url": "https://www.google.com/search?q=openai&hl=en&gl=us",
    "screenshot_key": "serp-scrapes/<request-id>/openai-<hash>.png",
    "content_key": "serp-scrapes/<request-id>/openai-<hash>.html",
    "data": {
        "sponsored_results": [
            {
                "kind": "sponsored_result",
                "index": 0,
                "title": "...",
                "brand": "...",
                "url": "https://...",
                "domain": "example.com"
            }
        ],
        "sponsored_products": [
            {
                "kind": "sponsored_product",
                "index": 0,
                "domain": "example.com",
                "url": "https://...",
                "image": "https://...",
                "title": "...",
                "price": "$19.99",
                "brand": "..."
            }
        ],
        "organic_results": [
            {
                "kind": "organic_result",
                "index": 0,
                "title": "...",
                "url": "https://...",
                "domain": "example.com"
            }
        ],
        "organic_products": [
            {
                "kind": "organic_product",
                "index": 0,
                "title": "...",
                "image": "https://..."
            }
        ],
        "similar_queries": [
            { "kind": "similar_query", "index": 0, "value": "..." }
        ]
    }
}
```

`data` groups the parsed SERP sections: `sponsored_results`, `sponsored_products`, `organic_results`, `organic_products`, and `similar_queries`. Each item's `url`/`domain` (where present) reflect the **final destination** the link resolves to, not Google's redirect/tracking link — see [How the parser works](#how-the-parser-works). Any section that fails to parse (e.g. an unexpected markup change) degrades independently to `[]` rather than failing the whole request.

### Saved artifacts (S3)

Every invocation uploads two objects to `bucket`/`prefix` under the same key stem (sanitized query + a short hash of the final URL): a full-page `.png` screenshot (`screenshot_key`) and the raw `.html` page content (`content_key`). This happens **unconditionally** — for real results and for blocked/CAPTCHA pages alike — since the block page itself is the most useful artifact for diagnosing _why_ a given proxy/session got flagged.

### Blocked responses

If Google serves its "unusual traffic" interstitial instead of a results page, the handler does **not** raise and does **not** change the response shape — it logs a warning to CloudWatch (`Google served a block/CAPTCHA page for query=...`) and still returns `content_key`/`screenshot_key` pointing at the saved interstitial page, plus an all-empty `data` (parsing isn't attempted against a block page), so you can inspect _why_ a given proxy/session got blocked without having to parse a Lambda error payload.

## How the parser works

Rather than hardcoding Google's current (obfuscated, rotating) class names, `_DATA_EXTRACT_JS` in `lambda_handler.py` looks for long-lived structural hooks:

- **Sponsored results**: elements matching `[data-pcu]`; the title is the first `<span>`'s text, the brand is the second `<span>`'s text truncated before the first `https://` substring (Google often renders "Brand https://displayed-url.com" as one string), and the link comes from the element's own `href`.
- **Organic results**: `h3[id]` elements; the title is the heading text and the link comes from the heading's parent `href` (Google wraps organic headings directly in the result's anchor).
- **Organic products**: `product-viewer-entrypoint` custom elements; the title comes from the inner `<div>`'s `aria-label`, falling back to the `<img>`'s `alt` text when that's empty, and the image is the `<img>`'s `src`.
- **Sponsored products**: elements matching `[data-dtld]` (the attribute itself is the pre-resolved domain); within `div.pla-unit-container`, the link/image come from `a.pla-unit-img-container-link`, and title/price/brand come from `div[data-call_grow_wiz_event='true']`, `div[aria-label]`, and `span[role='text']` respectively.
- **Similar (related) queries**: `data-q` attributes on `div[data-notify-expansion]` elements, plus link text inside `div#botstuff`; entries under 5 characters are filtered out (matches common noise like stray punctuation).

Each section is parsed independently and falls back to `[]` on failure rather than failing the whole extraction. Links (`href`s) captured by the JS above are frequently `/url?q=...`-style Google redirects rather than the destination itself, so `_final_url()` resolves each one to its final destination by issuing a real HTTP request through the same browser context/proxy as the search, with a small in-memory cache (`_URL_CACHE`, ~30 minute TTL) to avoid redundant lookups for links/destinations that repeat within or across nearby invocations.

A cookie-consent interstitial (shown for some EU-ish exit IPs regardless of `hl`/`gl`) is best-effort dismissed before parsing, by clicking `#L2AGLb` ("I agree") or an "Accept all"/"I agree" button if present.

## IAM permissions

In addition to `AWSLambdaBasicExecutionRole` (CloudWatch Logs), the execution role needs write access to the target bucket:

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

| Setting                    | Value                                                             | Why                                                                                                                          |
| -------------------------- | ----------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------- |
| Package type               | Image                                                             | Required — this is a container image, not a zip.                                                                             |
| Architecture               | `arm64`                                                           | Roughly 20% cheaper than x86_64. Match the architecture you built for.                                                       |
| Memory                     | 3008 MB                                                           | Memory in Lambda is tied to vCPU. Below ~1769 MB Chromium starts noticeably slower.                                          |
| Timeout                    | 60-90 s                                                           | Single search + parse is fast once loaded (a few seconds warm); budget for cold start (~80-90 s) plus proxy latency if used. |
| Ephemeral storage (`/tmp`) | 512 MB (default)                                                  | The screenshot and page content are streamed directly to S3, not accumulated on disk.                                        |
| Networking                 | Default (no VPC)                                                  | Binary is baked in; only outbound HTTPS to Google, your proxy (if any), and S3 is needed.                                    |
| Execution role             | Basic execution + S3 `PutObject` on the target bucket (see above) |                                                                                                                              |

## Cold start

Same profile as the sibling examples: first invocation in a new container takes ~80–90 s. Subsequent warm invocations are much faster.
