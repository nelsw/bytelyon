# CloakBrowser site crawler on AWS Lambda

Breadth-first crawl of a site from one or more seed URLs, using CloakBrowser's stealth Chromium inside an AWS Lambda function (container image package type). Every page visited is screenshotted and the screenshot is uploaded to S3; the function returns the list of URLs it crawled.

This is a sibling of [`../news-scraper`](../news-scraper), [`../page-scraper`](../page-scraper), [`../scraper`](../scraper), and [`../serp-scraper`](../serp-scraper) — all five build `FROM` the shared [`../base-image`](../base-image), which carries the Dockerfile/entrypoint scaffolding derived from the official `cloakhq/cloakbrowser` Docker Hub image (plus `boto3` for the S3 upload) that used to be duplicated in every handler's own `Dockerfile`. Every other invocation surface from the canonical image (`python`, `cloakserve`, `cloaktest`, `node`, `bash`, examples) keeps working.

This document does not prescribe a deployment method — push the resulting image to ECR and create the Lambda function however you prefer (AWS CLI, CDK, Terraform, SAM, console, etc.).

## Files in this directory

| File                | Purpose                                                                                                                                                                                           |
| ------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `Dockerfile`        | `FROM` the shared `../base-image`, plus a single `COPY lambda_handler.py`. See `../base-image/INSTRUCTIONS.md` for what the base image provides (including `boto3`, used here for the S3 upload). |
| `lambda_handler.py` | Crawl handler. Takes `{url or urls, bucket, ...}`, returns a list of crawled URLs. Headless by default.                                                                                           |
| `INSTRUCTIONS.md`   | This file.                                                                                                                                                                                        |

This directory is no longer fully standalone: it depends on the shared base image (`../base-image`) having been built and either published to ECR or loaded locally first — see `../base-image/INSTRUCTIONS.md`. Everything specific to this handler (its Dockerfile and `lambda_handler.py`) still lives entirely in this directory.

## Build

Build and publish `../base-image` first (see `../base-image/INSTRUCTIONS.md`), then:

```bash
docker buildx build --platform linux/arm64 --provenance=false --sbom=false \
  -t bytelyon-site-crawler:arm64 --load .
```

`--provenance=false --sbom=false` is required — without it, buildx produces a manifest-list/attestation image that Lambda's `CreateFunction` rejects with `InvalidParameterValueException: ... image manifest, config or layer media type ... is not supported`.

To build against a locally-loaded (not-yet-pushed) base image instead of pulling from ECR, add `--build-arg BASE_IMAGE=bytelyon-cloakbrowser-base:arm64`. For x86_64, switch `--platform linux/amd64` (and build the base image for the same architecture).

## Local smoke test (no Lambda deploy needed, but real AWS credentials required)

Screenshots are genuinely uploaded to S3 even in local RIE testing, so pass through real AWS credentials with S3 write access to the target bucket:

```bash
docker run --rm -p 9000:8080 \
  -e AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY -e AWS_SESSION_TOKEN -e AWS_REGION=us-east-1 \
  bytelyon-site-crawler:arm64

# In another shell:
curl -sS -XPOST "http://localhost:9000/2015-03-31/functions/function/invocations" \
  -d '{"url":"https://example.com","max_pages":5,"max_depth":1}'
```

Other invocation surfaces stay intact (these match the canonical CloakHQ image):

```bash
docker run --rm -it bytelyon-site-crawler:arm64 python                          # REPL
docker run --rm bytelyon-site-crawler:arm64 python examples/basic.py            # examples
docker run --rm -p 9222:9222 bytelyon-site-crawler:arm64 cloakserve --port=9222 # CDP server
docker run --rm bytelyon-site-crawler:arm64 cloaktest                           # stealth tests
docker run --rm -it bytelyon-site-crawler:arm64 node                            # JS wrapper
```

## Event schema

Exactly one of `url` / `urls` is required.

| Field                            | Type      | Default                                                                         |
| -------------------------------- | --------- | ------------------------------------------------------------------------------- |
| `url`                            | str       | a single seed URL                                                               |
| `urls`                           | list[str] | one or more seed URLs, checked before `url` if both present                     |
| `bucket`                         | str       | `"bytelyon-private"`                                                            |
| `prefix`                         | str       | `"crawls/{aws_request_id}/"` — S3 key prefix for this invocation's screenshots  |
| `max_pages`                      | int       | `20` — hard cap on total pages visited across all seeds                         |
| `max_depth`                      | int       | `2` — link-following depth from each seed (`0` = seeds only, no link-following) |
| `same_domain_only`               | bool      | `true` — only follow links whose hostname exactly matches the seed's hostname   |
| `full_page_screenshot`           | bool      | `true` — capture the entire scrollable page, not just the viewport              |
| `headless`                       | bool      | `true`                                                                          |
| `humanize`                       | bool      | `true`                                                                          |
| `human_preset`                   | str       | `"careful"`                                                                     |
| `goto_timeout_ms`                | int       | `30000`                                                                         |
| `wait_for_load_state_timeout_ms` | int       | `5000`                                                                          |

One browser + one page is launched per invocation and reused for every page in the crawl (sequential, not concurrent — simplest and cheapest for a single Lambda invocation).

### Crawl algorithm

Breadth-first from the seed(s):

1. Dequeue the next URL (fragment-stripped for dedup); skip if already visited.
2. Validate its scheme + resolve its hostname, rejecting private/internal targets (see Security below); navigate, wait, then re-validate the (possibly redirected) final URL.
3. Screenshot the page and upload it to `s3://{bucket}/{prefix}{sanitized-host-and-path}-{sha256[:10]}.png`.
4. Record the final URL in the result list.
5. If under `max_depth`, extract every `<a href>` on the page, keep the ones that are same-scheme, unvisited, and (if `same_domain_only`) share the seed's exact hostname, and enqueue them at `depth + 1`.
6. Repeat until the queue is empty, `max_pages` is reached, or the invocation is running low on time (see below).

### Response

```json
[
  "https://example.com/",
  "https://example.com/about",
  "https://example.com/contact"
]
```

Screenshots are a side effect written to S3 — they are **not** included in the response payload (unlike `../aws_lambda`'s one-shot handler, which returns `screenshot_b64` inline). Fetch them from S3 using the same key scheme, or list `s3://{bucket}/{prefix}` after the invocation completes.

## Lambda-specific additions

- **Chromium hardening (do not remove)**: `--disable-dev-shm-usage` and `--no-zygote`, same as the sibling examples.
- **URL validation, applied continuously, not just at the seed**: every discovered link is scheme- and SSRF-checked before navigation, and the final (possibly redirected) URL is re-checked after navigation — the same defense-in-depth as `../aws_lambda`, but more important here because a crawl follows links found _on the pages it visits_, which is attacker-influenced input if the crawled site is untrusted or compromised.
- **Time-budget-aware crawling**: before starting each new page, the handler checks `context.get_remaining_time_in_millis()` and stops enqueuing further pages once fewer than 10 seconds remain, returning whatever was crawled so far instead of getting killed mid-request by the Lambda platform. Set the function timeout generously relative to `max_pages` (see below) so this is a rare safety net, not the normal stopping condition.
- **`same_domain_only`**: matches hostname exactly (not a registrable-domain / public-suffix match), so `www.example.com` and `example.com` are treated as different hosts. Widen this yourself if you need subdomain-inclusive crawling.

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

| Setting                    | Value                                                             | Why                                                                                                                                                                                                                  |
| -------------------------- | ----------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Package type               | Image                                                             | Required — this is a container image, not a zip.                                                                                                                                                                     |
| Architecture               | `arm64`                                                           | Roughly 20% cheaper than x86_64. Match the architecture you built for.                                                                                                                                               |
| Memory                     | 3008 MB                                                           | Memory in Lambda is tied to vCPU. Below ~1769 MB Chromium starts noticeably slower.                                                                                                                                  |
| Timeout                    | scale with `max_pages`                                            | Budget ~10-15 s/page (nav + settle + screenshot + upload) plus ~10 s cold start. E.g. `max_pages=20` → 300-330 s; keep the default `max_pages=20` under a 300 s timeout, or lower `max_pages` for a shorter timeout. |
| Ephemeral storage (`/tmp`) | 512 MB (default)                                                  | Screenshots are streamed to S3, not accumulated on disk.                                                                                                                                                             |
| Networking                 | Default (no VPC)                                                  | Binary is baked in; only outbound HTTPS to the crawled sites and to S3 is needed.                                                                                                                                    |
| Execution role             | Basic execution + S3 `PutObject` on the target bucket (see above) |                                                                                                                                                                                                                      |

## Cold start

Same profile as the sibling examples: first invocation in a new container takes ~80–90 s. Subsequent warm invocations are much faster.

## License

The patched Chromium binary inside the upstream `cloakhq/cloakbrowser` image is governed by the **CloakBrowser Binary License** (published at https://github.com/CloakHQ/CloakBrowser/blob/main/BINARY-LICENSE.md). Internal organizational use (private ECR, your own scraping pipelines, your own business) is free. Exposing this Lambda as a paid API to third-party customers — i.e. browser-as-a-service — requires an OEM/SaaS license from CloakHQ (`cloakhq@pm.me`). Do not push the resulting image to a public registry; that would be redistribution and is prohibited.
