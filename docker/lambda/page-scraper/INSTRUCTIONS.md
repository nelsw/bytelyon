# CloakBrowser single-page scraper on AWS Lambda

Given a single URL: uploads a full-page screenshot to S3, and returns the page's links, meta tags, final URL, and title as JSON — using CloakBrowser's stealth Chromium inside an AWS Lambda function (container image package type).

This is a sibling of [`../news-scraper`](../news-scraper), [`../scraper`](../scraper), [`../serp-scraper`](../serp-scraper), and [`../site-crawler`](../site-crawler) — all five build `FROM` the shared [`../base-image`](../base-image), which carries the Dockerfile/entrypoint scaffolding derived from the official `cloakhq/cloakbrowser` Docker Hub image (plus `boto3` for the S3 upload) that used to be duplicated in every handler's own `Dockerfile`. Every other invocation surface from the canonical image (`python`, `cloakserve`, `cloaktest`, `node`, `bash`, examples) keeps working.

## Files in this directory

| File                | Purpose                                                                                                                                          |
| ------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------ |
| `Dockerfile`        | `FROM` the shared `../base-image`, plus a single `COPY lambda_handler.py`. See `../base-image/INSTRUCTIONS.md` for what the base image provides. |
| `lambda_handler.py` | Handler. Takes `{url, ...}`, returns `{url, title, links, meta, screenshot_s3_uri}`. Headless by default.                                        |
| `INSTRUCTIONS.md`   | This file.                                                                                                                                       |

This directory is no longer fully standalone: it depends on the shared base image (`../base-image`) having been built and either published to ECR or loaded locally first — see `../base-image/INSTRUCTIONS.md`. Everything specific to this handler (its Dockerfile and `lambda_handler.py`) still lives entirely in this directory.

## Build

Build and publish `../base-image` first (see `../base-image/INSTRUCTIONS.md`), then:

```bash
docker buildx build --platform linux/arm64 --provenance=false --sbom=false \
  -t bytelyon-page-scraper:arm64 --load .
```

`--provenance=false --sbom=false` is required — without it, buildx produces a manifest-list/attestation image that Lambda's `CreateFunction` rejects with `InvalidParameterValueException: ... image manifest, config or layer media type ... is not supported`.

To build against a locally-loaded (not-yet-pushed) base image instead of pulling from ECR:

```bash
docker buildx build --platform linux/arm64 --provenance=false --sbom=false \
  --build-arg BASE_IMAGE=bytelyon-cloakbrowser-base:arm64 \
  -t bytelyon-page-scraper:arm64 --load .
```

For x86_64, switch `--platform linux/amd64` (and build the base image for the same architecture).

## Local smoke test (no Lambda deploy needed, but real AWS credentials required)

The screenshot is genuinely uploaded to S3 even in local RIE testing, so pass through real AWS credentials with write access to the target bucket:

```bash
docker run --rm -p 9000:8080 \
  -e AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY -e AWS_SESSION_TOKEN -e AWS_REGION=us-east-1 \
  bytelyon-page-scraper:arm64

# In another shell:
curl -sS -XPOST "http://localhost:9000/2015-03-31/functions/function/invocations" \
  -d '{"url":"https://example.com"}'
```

Other invocation surfaces stay intact (these match the canonical CloakHQ image):

```bash
docker run --rm -it bytelyon-page-scraper:arm64 python                          # REPL
docker run --rm bytelyon-page-scraper:arm64 python examples/basic.py            # examples
docker run --rm -p 9222:9222 bytelyon-page-scraper:arm64 cloakserve --port=9222 # CDP server
docker run --rm bytelyon-page-scraper:arm64 cloaktest                           # stealth tests
docker run --rm -it bytelyon-page-scraper:arm64 node                            # JS wrapper
```

## Event schema

| Field                            | Type | Default                                                                                                                        |
| -------------------------------- | ---- | ------------------------------------------------------------------------------------------------------------------------------ |
| `url`                            | str  | required — the page to scrape (http/https only)                                                                                |
| `bucket`                         | str  | `"bytelyon-private"`                                                                                                           |
| `prefix`                         | str  | `"page-scrapes/{aws_request_id}/"` — S3 key prefix for the screenshot                                                          |
| `headless`                       | bool | `true`                                                                                                                         |
| `humanize`                       | bool | `true`                                                                                                                         |
| `human_preset`                   | str  | `"careful"`                                                                                                                    |
| `goto_timeout_ms`                | int  | `30000`                                                                                                                        |
| `wait_for_load_state_timeout_ms` | int  | `5000`                                                                                                                         |
| `include_links`                  | bool | `true` — when `false`, link extraction is skipped entirely (no DOM walk over anchors) and `links` is omitted from the response |

### Response

```json
{
  "url": "https://example.com/",
  "title": "Example Domain",
  "links": ["https://example.com/about", "https://example.com/contact"],
  "meta": [
    { "charset": "utf-8" },
    { "name": "description", "content": "..." },
    { "property": "og:title", "content": "..." }
  ],
  "screenshot_s3_uri": "s3://bytelyon-private/page-scrapes/<request-id>/example_com-<hash>.png"
}
```

- **`links`** — every `<a href>` on the page, resolved to absolute URLs, in DOM order, filtered to same-domain links only (the linked host must match the page's own host, ignoring a `www.` prefix on either side) with any link containing a URL fragment (`#...`) omitted entirely. Among what remains, `www.` and non-`www.` variants of the same URL (same scheme, path, and query, differing only by a `www.` host prefix) are treated as duplicates; the first-encountered form is kept. Pass `"include_links": false` to skip this extraction step altogether and omit `links` from the response.
- **`meta`** — every `<meta>` tag on the page, not a curated subset. Each entry includes whichever of `name` / `property` / `http-equiv` / `charset` is present on that tag, plus `content` if present. (For a curated OpenGraph/Twitter/news-style extraction instead, see `../news-scraper`.)
- **`url`** / **`title`** — the final (post-redirect) URL and `document.title`.
- **`screenshot_s3_uri`** — not part of the original ask, but included since otherwise there'd be no way to locate the uploaded screenshot from the response. Remove it from `lambda_handler.py`'s return dict if you'd rather keep the response to exactly the four requested fields.

## Lambda-specific additions

- **Chromium hardening (do not remove)**: `--disable-dev-shm-usage` and `--no-zygote`, same as the sibling examples.
- **URL validation**: scheme + SSRF checks on the input `url` before navigation, and again on `page.url` after navigation/redirects — same defense-in-depth as the other handlers that accept caller-supplied URLs.

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
| -------------------------- | ----------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------ |
| Package type               | Image                                                             | Required — this is a container image, not a zip.                                                             |
| Architecture               | `arm64`                                                           | Roughly 20% cheaper than x86_64. Match the architecture you built for.                                       |
| Memory                     | 3008 MB                                                           | Memory in Lambda is tied to vCPU. Below ~1769 MB Chromium starts noticeably slower.                          |
| Timeout                    | 60-90 s                                                           | Single page load + full-page screenshot + upload is fast once warm; budget for cold start (~80-90 s) on top. |
| Ephemeral storage (`/tmp`) | 512 MB (default)                                                  | The screenshot is streamed to S3, not accumulated on disk.                                                   |
| Networking                 | Default (no VPC)                                                  | Binary is baked in; only outbound HTTPS to the target site and to S3 is needed.                              |
| Execution role             | Basic execution + S3 `PutObject` on the target bucket (see above) |                                                                                                              |

## Cold start

Same profile as the sibling examples: first invocation in a new container takes ~80–90 s. Subsequent warm invocations are much faster.

