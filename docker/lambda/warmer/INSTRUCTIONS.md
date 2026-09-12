# CloakBrowser base-image warmer on AWS Lambda

A single-purpose Lambda function that launches (and immediately closes) a CloakBrowser browser instance and returns some timing/version diagnostics — nothing else. It does not scrape, navigate anywhere meaningful, write to S3, or accept a proxy. Its only job is to exercise the expensive part of a cold start (baked Chromium binary extraction/verification, renderer process spawn, CloakBrowser's stealth patch injection) so you can either:

1. Point a periodic trigger (an EventBridge scheduled rule) at it to keep its own execution environment warm, or
2. Invoke it once after building/pushing a new base image as a fast, cheap, side-effect-free smoke test that it still boots.

This is a sibling of [`../news-scraper`](../news-scraper), [`../page-scraper`](../page-scraper), [`../scraper`](../scraper), [`../serp-scraper`](../serp-scraper), and [`../site-crawler`](../site-crawler) — same shared [`../base-image`](../base-image), different (much smaller) `lambda_handler.py`.

## Read this before wiring up a scheduled trigger

**Invoking this function does NOT warm any _other_ Lambda function.** Even functions built from the byte-identical container image get their own independent pool of execution environments — AWS Lambda does not share warm containers across function ARNs, regardless of shared image layers. Pinging this function only keeps _this function's own_ containers warm. If your actual goal is to stop `../serp-scraper` (or any other handler) from going cold, your options are:

- **AWS Lambda Provisioned Concurrency** on that handler's own function — the correct, AWS-native fix, though it has an hourly cost regardless of invocations.
- **A scheduled ping of that handler's own function** with a cheap-but-real event (e.g. `{"query": "test"}` for `serp-scraper`) — simple, but runs that handler's actual logic (and its side effects, like S3 writes) on every ping, and consumes its normal IAM permissions/proxy config.
- **This function**, if what you actually want is: a function with no side effects and no extra IAM permissions that proves the shared base image still boots, and/or you're fine with only _this_ function itself staying warm (e.g. you're about to fan out a burst of concurrent invocations across the other handlers and want to prime the account's Lambda container-image pull/cache path first — image-layer pull caching within a region/account is an internal Lambda implementation detail, not a documented guarantee, so treat this as a best-effort nudge, not a fix).

## Files in this directory

| File                | Purpose                                                                                                         |
| ------------------- | --------------------------------------------------------------------------------------------------------------- |
| `Dockerfile`        | `FROM` the shared `../base-image`, plus a single `COPY lambda_handler.py`. See `../base-image/INSTRUCTIONS.md`. |
| `lambda_handler.py` | Launch + close. No scraping, no S3, no proxy support. Headless by default.                                      |
| `INSTRUCTIONS.md`   | This file.                                                                                                      |

This directory is no longer fully standalone: it depends on the shared base image (`../base-image`) having been built and either published to ECR or loaded locally first — see `../base-image/INSTRUCTIONS.md`.

## Build

Build and publish `../base-image` first (see `../base-image/INSTRUCTIONS.md`), then:

```bash
docker buildx build --platform linux/arm64 --provenance=false --sbom=false \
  -t bytelyon-warmer:arm64 --load .
```

`--provenance=false --sbom=false` is required — without it, buildx produces a manifest-list/attestation image that Lambda's `CreateFunction` rejects with `InvalidParameterValueException: ... image manifest, config or layer media type ... is not supported`.

To build against a locally-loaded (not-yet-pushed) base image instead of pulling from ECR, add `--build-arg BASE_IMAGE=bytelyon-cloakbrowser-base:arm64`. For x86_64, switch `--platform linux/amd64` (and build the base image for the same architecture).

## Local smoke test (no AWS account needed)

```bash
docker run --rm -p 9000:8080 bytelyon-warmer:arm64

# In another shell — no payload needed at all:
curl -sS -XPOST "http://localhost:9000/2015-03-31/functions/function/invocations" -d '{}'

# Invoke again on the same container to see cold_start flip to false:
curl -sS -XPOST "http://localhost:9000/2015-03-31/functions/function/invocations" -d '{}'

# Also exercise navigation (not just process startup):
curl -sS -XPOST "http://localhost:9000/2015-03-31/functions/function/invocations" \
  -d '{"navigate": true}'
```

No AWS credentials, S3 bucket, or proxy are needed for this handler at all — that's deliberate, so it can't fail for reasons unrelated to "does the base image still boot".

## Event schema

| Field      | Type | Default | Description                                                                                                                    |
| ---------- | ---- | ------- | ------------------------------------------------------------------------------------------------------------------------------ |
| `headless` | bool | `true`  |                                                                                                                                |
| `navigate` | bool | `false` | Also open a page and navigate it to `about:blank` (no network call) to exercise the render pipeline, not just process startup. |

### Response

```json
{
  "warm": true,
  "cold_start": true,
  "cloakbrowser_version": "0.5.10",
  "chromium_version": "146.0.7680.177.5",
  "launch_ms": 812.4,
  "navigate_ms": null,
  "total_ms": 815.9
}
```

`cold_start` is `true` only on a given container's first invocation (tracked with a module-level flag, so it resets whenever Lambda spins up a fresh execution environment) — on later invocations of an already-warm container it's `false`. If you're pinging this function on a schedule specifically to keep it warm, seeing `cold_start: true` on a ping means the container was recycled since the last one (Lambda reclaims idle containers, typically well within an hour of inactivity) — check your trigger's interval if that's happening more than expected. `navigate_ms` is `null` unless `"navigate": true` was passed.

## IAM permissions

Just `AWSLambdaBasicExecutionRole` (CloudWatch Logs) — nothing else. This handler never touches S3, a proxy, or any caller-supplied URL, so it needs no additional grants.

## Scheduling a periodic warm-up (optional)

An EventBridge scheduled rule targeting this function's ARN, e.g. every 5 minutes:

```bash
aws events put-rule --name warm-cloakbrowser --schedule-expression "rate(5 minutes)"
aws lambda add-permission --function-name bytelyon-warmer \
  --statement-id AllowEventBridge --action lambda:InvokeFunction \
  --principal events.amazonaws.com --source-arn <rule-arn>
aws events put-targets --rule warm-cloakbrowser \
  --targets "Id"="1","Arn"="<function-arn>","Input"="{}"
```

Remember: this only keeps _this_ function's own containers warm (see the warning above) — it is not a substitute for Provisioned Concurrency on a handler whose cold start latency actually matters to you.

## Function configuration recommendations

| Setting                          | Value                                       | Why                                                                                                                                                             |
| -------------------------------- | ------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Package type                     | Image                                       | Required — this is a container image, not a zip.                                                                                                                |
| Architecture                     | `arm64`                                     | Roughly 20% cheaper than x86_64. Match the architecture you built for.                                                                                          |
| Memory                           | 3008 MB                                     | Same as every other handler here — Chromium still has to actually launch. Below ~1769 MB it starts noticeably slower, which would defeat the point of a warmer. |
| Timeout                          | 30 s                                        | A launch+close with no navigation completes in well under a second warm, a few seconds on a genuine cold start. 30 s leaves generous headroom.                  |
| Ephemeral storage (`/tmp`)       | 512 MB (default)                            | Nothing is written to disk.                                                                                                                                     |
| Networking                       | Default (no VPC)                            | No outbound network call at all unless `"navigate": true` (and even then, `about:blank` needs none).                                                            |
| Execution role                   | `AWSLambdaBasicExecutionRole`               | Just CloudWatch Logs — see IAM permissions above.                                                                                                               |
| Reserved/provisioned concurrency | none needed for this function's own purpose | Only relevant if you're using _this_ function as a literal warm pool rather than just a smoke test.                                                             |

## Cold start

Same profile as the sibling examples: first invocation in a new container takes ~80–90 s (baked Chromium binary extraction/verification is the dominant cost). Subsequent warm invocations complete in well under a second, since there's no navigation or I/O by default.

## License

Same as every sibling handler: the patched Chromium binary inside the upstream `cloakhq/cloakbrowser` image is governed by the **CloakBrowser Binary License** (published at https://github.com/CloakHQ/CloakBrowser/blob/main/BINARY-LICENSE.md). Internal organizational use (private ECR, your own scraping pipelines, your own business) is free. Do not push the resulting image to a public registry; that would be redistribution and is prohibited.
