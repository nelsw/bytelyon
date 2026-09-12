# CloakBrowser news/article metadata scraper on AWS Lambda

Batch-extract article body text and OpenGraph/Twitter/news metadata (image, description, keywords) from one or more URLs per invocation, using CloakBrowser's stealth Chromium inside an AWS Lambda function (container image package type).

This is a sibling of [`../page-scraper`](../page-scraper), [`../scraper`](../scraper), [`../serp-scraper`](../serp-scraper), and [`../site-crawler`](../site-crawler) — all five build `FROM` the shared [`../base-image`](../base-image), which carries the Dockerfile/entrypoint scaffolding derived from the official `cloakhq/cloakbrowser` Docker Hub image that used to be duplicated in every handler's own `Dockerfile`. This handler is ported from a standalone script that scraped a list of URLs given on the command line. Every other invocation surface from the canonical image (`python`, `cloakserve`, `cloaktest`, `node`, `bash`, examples) keeps working, same as the sibling handlers.

This document does not prescribe a deployment method — push the resulting image to ECR and create the Lambda function however you prefer (AWS CLI, CDK, Terraform, SAM, console, etc.).

## Files in this directory

| File                | Purpose                                                                                                                                                                     |
| ------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `Dockerfile`        | `FROM` the shared `../base-image`, plus a single `COPY lambda_handler.py`. See `../base-image/INSTRUCTIONS.md` for what the base image provides.                            |
| `lambda_handler.py` | Batch handler. Takes `{url or urls, ...}`, returns a list of `{url, body, img_src, img_alt, description, keywords}`. Headless by default (no Xvfb needed for this handler). |
| `INSTRUCTIONS.md`   | This file.                                                                                                                                                                  |

This directory is no longer fully standalone: it depends on the shared base image (`../base-image`) having been built and either published to ECR or loaded locally first — see `../base-image/INSTRUCTIONS.md`. Everything specific to this handler (its Dockerfile and `lambda_handler.py`) still lives entirely in this directory.

## Build

Build and publish `../base-image` first (see `../base-image/INSTRUCTIONS.md`), then, from inside this directory:

```bash
docker buildx build --platform linux/arm64 --provenance=false --sbom=false \
  -t bytelyon-news-scraper:arm64 --load .
```

`--provenance=false --sbom=false` is required — without it, buildx produces a manifest-list/attestation image that Lambda's `CreateFunction` rejects with `InvalidParameterValueException: ... image manifest, config or layer media type ... is not supported`.

Or from anywhere, pointing at this directory as the build context:

```bash
docker buildx build --platform linux/arm64 --provenance=false --sbom=false \
  -f path/to/news-scraper/Dockerfile \
  -t bytelyon-news-scraper:arm64 --load \
  path/to/news-scraper
```

To build against a locally-loaded (not-yet-pushed) base image instead of pulling from ECR, add `--build-arg BASE_IMAGE=bytelyon-cloakbrowser-base:arm64`. For x86_64, switch `--platform linux/amd64` (and build the base image for the same architecture).

## Local smoke test (no AWS account needed)

The image bakes in `aws-lambda-rie`, so the standard Lambda local-invoke endpoint works without mounting anything:

```bash
docker run --rm -p 9000:8080 bytelyon-news-scraper:arm64

# In another shell:
curl -sS -XPOST "http://localhost:9000/2015-03-31/functions/function/invocations" \
  -d '{"urls":["https://example.com"]}'
```

Other invocation surfaces stay intact (these match the canonical CloakHQ image):

```bash
docker run --rm -it bytelyon-news-scraper:arm64 python                          # REPL
docker run --rm bytelyon-news-scraper:arm64 python examples/basic.py            # examples
docker run --rm -p 9222:9222 bytelyon-news-scraper:arm64 cloakserve --port=9222 # CDP server
docker run --rm bytelyon-news-scraper:arm64 cloaktest                           # stealth tests
docker run --rm -it bytelyon-news-scraper:arm64 node                            # JS wrapper
```

## Event schema

Exactly one of `url` / `urls` is required.

| Field                            | Type      | Default                                                           |
| -------------------------------- | --------- | ----------------------------------------------------------------- |
| `url`                            | str       | a single URL to process                                           |
| `urls`                           | list[str] | one or more URLs to process, checked before `url` if both present |
| `headless`                       | bool      | `true` — matches the source script                                |
| `humanize`                       | bool      | `true` — matches the source script                                |
| `human_preset`                   | str       | `"careful"` — matches the source script                           |
| `goto_timeout_ms`                | int       | `30000` — Playwright default; the source script left this unset   |
| `wait_for_load_state_timeout_ms` | int       | `5000` — matches the source script's hardcoded `5_000`            |

One browser + one page is launched per invocation and reused across every URL in `urls`, in order — same as the source script's loop over `sys.argv[1:]`.

### Response

A JSON array, one object per URL, in the same order as the request:

```json
[
  {
    "url": "https://example.com",
    "body": "...",
    "img_src": "...",
    "img_alt": "...",
    "description": "...",
    "keywords": ["..."]
  }
]
```

If navigation/extraction for a given URL raises anything other than a Playwright timeout (the source script already tolerates timeouts and extracts whatever loaded), that URL's entry becomes `{"url": ..., "error": "..."}` instead of failing the whole batch — this is the one behavioral addition beyond a literal port, so one bad URL in a multi-URL request doesn't discard results already gathered for the others.

## Ported from the source script — behavior notes

The handler is a close port of a CLI script that did:

```python
browser = launch(headless=True, humanize=True, human_preset="careful")
page = browser.new_page()
for url in sys.argv[1:]:
    page.goto(url)
    page.wait_for_load_state('domcontentloaded', timeout=5_000)
    data.append({...})
print(json.dumps(data, indent=2))
```

Two quirks from the source script are **preserved as-is** (not fixed, since the ask was to reproduce its result):

- `body()`'s selector fallback loop (`for selector in ['article', 'main', 'body', 'html']: locator = page.locator(selector); if locator is not None: break`) always breaks on the first iteration — `page.locator(...)` returns a lazy handle and is never `None`, even when nothing matches. In practice `body()` always resolves against the `article` selector and never falls back to `main`/`body`/`html`. If a page has no `<article>` element, `body` comes back empty.
- The source script's `if txt is not "":` was changed to `if txt != "":` here — the former is an identity comparison against a string literal, which raises a `SyntaxWarning` on modern Python and isn't guaranteed to behave consistently across interpreters. Behavior is unchanged in practice (CPython interns short string literals).

## Lambda-specific additions (not in the source script)

The source script assumes a persistent local process with a trusted, operator-supplied URL list. Running the same logic as an internet-reachable-ish Lambda function warrants a couple of defensive additions, mirroring `../aws_lambda`:

- **Lambda Chromium hardening (do not remove)**: `--disable-dev-shm-usage` (Lambda's `/dev/shm` is ~64 MB) and `--no-zygote` (Lambda's restricted process model can't fork from Chromium's zygote process) are always passed to `launch()`.
- **URL validation** — scheme restricted to `http://`/`https://`, and hostnames are resolved and checked against private/loopback/link-local/reserved/multicast IP ranges before navigation (blocks SSRF to things like `169.254.169.254`) and re-checked against `page.url` after navigation to catch server-side redirects. See the **Security** section of `../aws_lambda/INSTRUCTIONS.md` for the same caveats (this doesn't stop a GET with side effects from reaching an internal endpoint before the response is discarded; DNS rebinding can theoretically bypass the pre-navigation check).

No retry orchestration is implemented (unlike `../aws_lambda`) — this keeps the handler a faithful, minimal adaptation of the source script rather than a redesign. Add it yourself if flaky navigations are a problem for your target sites.

## Function configuration recommendations

Whatever tool you use to create the Lambda function (CLI, CDK, Terraform, SAM, console), apply these settings:

| Setting                    | Value                         | Why                                                                                                                                                                                                                                                     |
| -------------------------- | ----------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Package type               | Image                         | Required — this is a container image, not a zip.                                                                                                                                                                                                        |
| Architecture               | `arm64`                       | Roughly 20% cheaper than x86_64. Native build on Apple Silicon. Match the architecture you built for.                                                                                                                                                   |
| Memory                     | 3008 MB                       | Memory in Lambda is tied to vCPU. Below ~1769 MB Chromium starts noticeably slower, even headless.                                                                                                                                                      |
| Timeout                    | scale with URL count          | Each URL can take up to `goto_timeout_ms` (default 30 s) + `wait_for_load_state_timeout_ms` (default 5 s) worst case, sequentially. E.g. for up to 5 URLs per invocation, budget ~200 s; for a single URL, 60-90 s is comfortable including cold start. |
| Ephemeral storage (`/tmp`) | 512 MB (default)              | No screenshots are captured here, so the 512 MB default is normally enough headroom for Chromium's profile dir.                                                                                                                                         |
| Networking                 | Default (no VPC)              | Binary is baked in, no network needed at cold start.                                                                                                                                                                                                    |
| Execution role             | `AWSLambdaBasicExecutionRole` | Just CloudWatch Logs. Add more permissions only if your handler needs them.                                                                                                                                                                             |

## Cold start

Same profile as `../aws_lambda`: first invocation in a new container takes ~80–90 s (image extraction, Chromium binary mmap, JS engine warmup). Subsequent warm invocations on the same container are much faster (page-load-bound, not init-bound).


