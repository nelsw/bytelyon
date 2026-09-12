# CloakBrowser fast single-page capture on AWS Lambda

Given a single URL (with an optional proxy): uploads the rendered HTML and a full-page screenshot to S3, and returns just the two S3 keys — using CloakBrowser's stealth Chromium inside an AWS Lambda function (container image package type).

This handler is deliberately minimal and optimized for **speed**: no link/meta extraction, no consent-dialog handling, no retry orchestration, no headed/Xvfb mode. It's a sibling of [`../news-scraper`](../news-scraper), [`../page-scraper`](../page-scraper), [`../scraper`](../scraper), [`../serp-scraper`](../serp-scraper), and [`../site-crawler`](../site-crawler) — all six build `FROM` the shared [`../base-image`](../base-image), which carries the Dockerfile/entrypoint scaffolding derived from the official `cloakhq/cloakbrowser` Docker Hub image (plus `boto3` for the S3 uploads) that used to be duplicated in every handler's own `Dockerfile`. Every other invocation surface from the canonical image (`python`, `cloakserve`, `cloaktest`, `node`, `bash`, examples) keeps working.

If you need link/meta extraction as well, see [`../page-scraper`](../page-scraper) instead — this handler trades that off for lower latency.

## Files in this directory

| File                | Purpose                                                                                                               |
| ------------------- | ----------------------------------------------------------------------------------------------------------------------|
| `Dockerfile`        | `FROM` the shared `../base-image`, plus a single `COPY lambda_handler.py`. See `../base-image/INSTRUCTIONS.md`.       |
| `lambda_handler.py` | Handler. Takes `{url, proxy?, ...}`, returns `{url, content_key, screenshot_key}`. Headless by default.               |
| `INSTRUCTIONS.md`   | This file.                                                                                                             |

This directory is no longer fully standalone: it depends on the shared base image (`../base-image`) having been built and either published to ECR or loaded locally first — see `../base-image/INSTRUCTIONS.md`. Everything specific to this handler (its Dockerfile and `lambda_handler.py`) still lives entirely in this directory.

## Build

Build and publish `../base-image` first (see `../base-image/INSTRUCTIONS.md`), then:

```bash
docker buildx build --platform linux/arm64 --provenance=false --sbom=false \
  -t bytelyon-grab:arm64 --load .
```

`--provenance=false --sbom=false` is required — without it, buildx produces a manifest-list/attestation image that Lambda's `CreateFunction` rejects with `InvalidParameterValueException: ... image manifest, config or layer media type ... is not supported`.

To build against a locally-loaded (not-yet-pushed) base image instead of pulling from ECR:

```bash
docker buildx build --platform linux/arm64 --provenance=false --sbom=false \
  --build-arg BASE_IMAGE=bytelyon-cloakbrowser-base:arm64 \
  -t bytelyon-grab:arm64 --load .
```

For x86_64, switch `--platform linux/amd64` (and build the base image for the same architecture).

## Local smoke test (no Lambda deploy needed, but real AWS credentials required)

Both artifacts are genuinely uploaded to S3 even in local RIE testing, so pass through real AWS credentials with write access to the target bucket:

```bash
docker run --rm -p 9000:8080 \
  -e AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY -e AWS_SESSION_TOKEN -e AWS_REGION=us-east-1 \
  bytelyon-grab:arm64

# In another shell:
curl -sS -XPOST "http://localhost:9000/2015-03-31/functions/function/invocations" \
  -d '{"url":"https://example.com"}'

# With a proxy:
curl -sS -XPOST "http://localhost:9000/2015-03-31/functions/function/invocations" \
  -d '{"url":"https://example.com", "proxy":"http://user:pass@host:port", "geoip":true}'
```

Other invocation surfaces stay intact (these match the canonical CloakHQ image):

```bash
docker run --rm -it bytelyon-grab:arm64 python                          # REPL
docker run --rm bytelyon-grab:arm64 python examples/basic.py            # examples
docker run --rm -p 9222:9222 bytelyon-grab:arm64 cloakserve --port=9222 # CDP server
docker run --rm bytelyon-grab:arm64 cloaktest                           # stealth tests
docker run --rm -it bytelyon-grab:arm64 node                            # JS wrapper
```

## Event schema

| Field                  | Type       | Default                                                                                                 |
| ---------------------- | ---------- | ----------------------------------------------------------------------------------------------------------------------------- |
| `url`                  | str        | required — the page to capture (http/https only)                                                       |
| `proxy`                | str\|dict  | none — `"http://user:pass@host:port"` (credentials split out automatically) or a `ProxySettings` dict   |
| `geoip`                | bool       | `false` — auto-derive timezone/locale from the proxy's exit IP (only meaningful with `proxy` set)        |
| `bucket`               | str        | `"bytelyon-private"`                                                                                     |
| `prefix`               | str        | `"grabs/{aws_request_id}/"` — S3 key prefix for both uploads                                          |
| `headless`             | bool       | `true`                                                                                                    |
| `humanize`             | bool       | `false` — off by default for speed; set `true` for sites that fingerprint mouse/keyboard/scroll behavior |
| `human_preset`         | str        | `"careful"`                                                                                               |
| `wait_until`           | str        | `"domcontentloaded"` — passed to `page.goto` (`"load"`\|`"domcontentloaded"`\|`"networkidle"`\|`"commit"`) |
| `goto_timeout_ms`      | int        | `30000`                                                                                                   |
| `full_page_screenshot` | bool       | `true`                                                                                                    |

### Response

```json
{
  "url": "https://example.com/",
  "content_key": "grabs/<request-id>/example_com-<hash>.html",
  "screenshot_key": "grabs/<request-id>/example_com-<hash>.png"
}
```

- **`content_key`** / **`screenshot_key`** — S3 keys (under `bucket`) for the rendered HTML (`page.content()`) and the screenshot, uploaded **concurrently** to shave the round-trip time roughly in half.
- **`url`** — the final (post-redirect) URL.

## Speed choices (this handler's reason to exist)

- **`humanize` defaults to `false`** — no synthetic mouse/keyboard/scroll delay. The sibling handlers default this to `true` for stealth; here it's opt-in, since the ask was raw speed. Set it to `true` per-invocation for sites that fingerprint interaction behavior.
- **`wait_until` defaults to `"domcontentloaded"`** — fires as soon as the DOM is parsed, without waiting for every subresource (images, fonts, analytics beacons) like `"load"` or `"networkidle"` would.
- **No link/meta extraction, no consent-dialog handling, no retry orchestration** — every extra step the sibling handlers take is left out. Best-effort on navigation timeout: whatever loaded so far still gets captured and uploaded rather than failing the whole invocation.
- **Concurrent S3 uploads** — HTML and screenshot are two independent network calls, submitted to a 2-worker thread pool instead of run back-to-back.

## Lambda-specific additions

- **Chromium hardening (do not remove)**: `--disable-dev-shm-usage` and `--no-zygote`, same as the sibling handlers.
- **URL validation**: scheme + SSRF checks on the input `url` before navigation, and again on `page.url` after navigation/redirects.
- **Proxy TCP preflight**: if `proxy` is set, a raw TCP connect to the proxy's host:port is attempted first (~8s timeout) so an unreachable/blocking proxy fails fast with a clear error instead of hanging until the function timeout. Same as `../serp-scraper`.

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

| Setting                    | Value                                                             | Why                                                                                                          |
| --------------------------- | ------------------------------------------------------------------ | -------------------------------------------------------------------------------------------------------------|
| Package type               | Image                                                             | Required — this is a container image, not a zip.                                                             |
| Architecture               | `arm64`                                                           | Roughly 20% cheaper than x86_64. Match the architecture you built for.                                       |
| Memory                     | 3008 MB                                                           | Memory in Lambda is tied to vCPU. Below ~1769 MB Chromium starts noticeably slower.                          |
| Timeout                    | 30-45 s                                                           | Single page load + full-page screenshot + concurrent uploads is fast once warm; budget for cold start (~80-90 s) on top. |
| Ephemeral storage (`/tmp`) | 512 MB (default)                                                  | Both artifacts are held in memory and streamed to S3, not accumulated on disk.                                |
| Networking                 | Default (no VPC)                                                  | Binary is baked in; only outbound HTTPS to the target site (or proxy) and to S3 is needed.                    |
| Execution role             | Basic execution + S3 `PutObject` on the target bucket (see above) |                                                                                                                |

## Cold start

Same profile as the sibling examples: first invocation in a new container takes ~80–90 s. Subsequent warm invocations are much faster — this handler is intentionally the leanest of the family, so warm-invocation latency should be at or below the other handlers'.
