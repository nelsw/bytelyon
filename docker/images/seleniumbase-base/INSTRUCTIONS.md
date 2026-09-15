# Shared SeleniumBase Lambda base image

The [SeleniumBase](https://seleniumbase.io/) counterpart of
[`../../lambda/base-image`](../../lambda/base-image): a shared AWS Lambda
base image that every `*-seleniumbase` handler builds `FROM`, published to
its own ECR repository. Where the CloakBrowser base image layers Lambda glue
on top of the official `cloakhq/cloakbrowser` Docker Hub image, this one is
built from scratch (a plain Python image + system Chromium), since there's
no official pre-built SeleniumBase image to start from.

## Why SeleniumBase instead of CloakBrowser here

SeleniumBase provides its own stealth browser automation stack (CDP Mode —
see the [SeleniumBase README](https://seleniumbase.io/)), and its **Stealthy
Playwright Mode** lets a normal `playwright.sync_api` session drive that
already-stealthy browser:

1. `seleniumbase.sb_cdp.Chrome(...)` launches SeleniumBase's patched,
   hardened Chromium directly (CDP Mode — no chromedriver/webdriver binary
   involved at all).
2. `sb.get_endpoint_url()` returns the browser's CDP endpoint.
3. `playwright.sync_api.sync_playwright().chromium.connect_over_cdp(...)`
   attaches an ordinary Playwright `Browser` to that already-running
   session.

[`seleniumbase_playwright.py`](seleniumbase_playwright.py) (baked into this
image) wraps that three-step dance behind a `launch()` function with the
same call shape as `cloakbrowser.launch()`, so a handler built `FROM` this
image is a near line-for-line port of its CloakBrowser-based counterpart:
only the `launch` import changes, every `page.goto()` / `page.content()` /
`page.screenshot()` call is identical, because both providers hand back a
standard Playwright `Page`. See the module docstring in
`seleniumbase_playwright.py` for the handful of interface differences
(`humanize`/`human_preset`/`geoip` are accepted but no-ops here).

## Why a base image at all

Same reasoning as [`../../lambda/base-image`](../../lambda/base-image):

- **Smaller deployment packages** — a handler rebuild/push only needs to
  transfer its own thin `lambda_handler.py` COPY layer; the shared layers
  (system Chromium, awslambdaric, the RIE, `seleniumbase` + `playwright`,
  the entrypoint, this glue module) already live in ECR and are reused via
  Docker's layer cache.
- **Reusability** — one Dockerfile to patch (a CVE fix, an entrypoint
  tweak, a SeleniumBase/Playwright version bump) instead of one per
  handler.
- **Faster deployments** — once the base image is in ECR, a handler
  build/push touches ~1 new layer instead of ~10.

See [`../../lambda/base-image/INSTRUCTIONS.md`](../../lambda/base-image/INSTRUCTIONS.md)'s
"A note on AWS Lambda Layers" section — it applies identically here (every
handler in this repo is `PackageType: Image`, so the native Lambda `Layers`
API field doesn't apply; a shared base image is the container-image
equivalent).

## Files in this directory

| File                       | Purpose                                                                                       |
| -------------------------- | ---------------------------------------------------------------------------------------------- |
| `Dockerfile`                | Plain Python image + system Chromium + all shared Lambda glue. No handler code.                |
| `lambda-entrypoint.sh`      | Dual-mode entrypoint. Verbatim copy of the CloakBrowser base image's — the logic is generic.    |
| `seleniumbase_playwright.py`| The `launch()` shim described above. Installed to `/opt/python` (on `PYTHONPATH`) in the image. |
| `INSTRUCTIONS.md`           | This file.                                                                                      |

## System Chromium, not Google Chrome

This image installs Debian's `chromium` apt package rather than Google
Chrome. Google only publishes **amd64** Linux builds of Chrome, and this
image — like every other image in this repo — targets **arm64** as well.
Debian's `chromium` package is available for both architectures and is
auto-discovered on `PATH` by SeleniumBase's `find_chrome_executable()`, so
no `browser_executable_path=` override is needed anywhere in
`seleniumbase_playwright.py` or handler code.

No `chromium-driver` package is installed: CDP Mode never touches
chromedriver, so it isn't needed. This means classic
Selenium/WebDriver-mode SeleniumBase usage (`Driver()`, `BaseCase`, `SB()`
without CDP Mode) is **not** supported by this image — only CDP Mode /
Stealthy Playwright Mode, which is all any handler built from it uses.

## Build & push

Do this before (re)building any `*-seleniumbase` handler image — handler
Dockerfiles default to pulling this image by tag from ECR.

```bash
aws ecr get-login-password --region us-east-1 \
  | docker login --username AWS --password-stdin <account>.dkr.ecr.us-east-1.amazonaws.com

docker buildx build --platform linux/arm64 --provenance=false --sbom=false \
  -t <account>.dkr.ecr.us-east-1.amazonaws.com/bytelyon-seleniumbase-base:latest \
  --push .
```

`--provenance=false --sbom=false` is required — without it, buildx produces
a manifest-list/attestation image that Lambda's `CreateFunction`/
`UpdateFunctionCode` rejects with `InvalidParameterValueException: ... image
manifest, config or layer media type ... is not supported`. This applies
transitively: every handler image built `FROM` this one must also be built
with these flags, every time.

For x86_64, switch `--platform linux/amd64` (and rebuild every handler for
the same architecture — mixing architectures between the base and a handler
will fail at Lambda invocation time, not at build time).

### Local-only testing (skip the ECR push)

```bash
docker buildx build --platform linux/arm64 --provenance=false --sbom=false \
  -t bytelyon-seleniumbase-base:arm64 --load .
```

Then point a handler at it without touching ECR:

```bash
docker buildx build --platform linux/arm64 --provenance=false --sbom=false \
  --build-arg BASE_IMAGE=bytelyon-seleniumbase-base:arm64 \
  -t bytelyon-grab-seleniumbase:arm64 --load ../../lambda/grab-seleniumbase
```

## Versioning

Same policy as [`../../lambda/base-image`](../../lambda/base-image): prefer
pushing a new tag (e.g. `:v2`) over overwriting `:latest` in place, then
bump the `BASE_IMAGE` default `ARG` in every handler `Dockerfile` you want
to pick up the change. Already-deployed Lambda functions are pinned to the
image **digest** Lambda resolved at the most recent `update-function-code`
call — moving `:latest` in ECR doesn't affect them until each is explicitly
redeployed.
