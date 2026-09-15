# Shared CloakBrowser Lambda base image

The common scaffolding previously duplicated verbatim across every handler's
own `Dockerfile` (`../news-scraper`, `../page-scraper`, `../scraper`,
`../serp-scraper`, `../site-crawler`) — the Lambda Runtime Interface Client,
the Lambda Runtime Interface Emulator, boto3, the dual-mode entrypoint, the
non-root readability fix, and the runtime `ENV` — now lives here, in a single
image, published to its own ECR repository. Every handler's `Dockerfile`
builds `FROM` it and adds only its own `lambda_handler.py`.

## Why

- **Smaller deployment packages** — a handler rebuild/push only needs to
  transfer the thin per-handler layer (a single `lambda_handler.py` COPY);
  the shared layers already live in ECR and are reused via Docker's layer
  cache, instead of every handler independently re-installing
  `awslambdaric`/`boto3`, re-downloading the RIE binary, etc.
- **Reusability** — one Dockerfile to patch when the Lambda glue needs to
  change (a CVE fix in `awslambdaric`, an entrypoint tweak, a new baked-in
  dependency), instead of five that can silently drift apart.
- **Faster deployments** — `docker build`/`push` for a handler after the base
  image is already in ECR touches ~1 new layer instead of ~10.

## Files in this directory

| File                   | Purpose                                                                                                |
| ---------------------- | ------------------------------------------------------------------------------------------------------ |
| `Dockerfile`           | `FROM cloakhq/cloakbrowser` plus all shared Lambda glue. No handler code, no `lambda_handler.py` COPY. |
| `lambda-entrypoint.sh` | Dual-mode entrypoint, identical to what every handler used to carry individually.                      |
| `fonts/`               | Empty by default (just its own README) — optional real Windows/Office fonts, see below.                |
| `INSTRUCTIONS.md`      | This file.                                                                                             |

## Windows font spoofing

CloakBrowser warns (`[cloakbrowser] Incomplete Windows font set ...`) when spoofing a Windows fingerprint without a full Windows font set installed — a bare Linux font list is itself a detection signal. The `Dockerfile` installs `ttf-mscorefonts-installer` (Debian's packaging of Microsoft's real "Core Fonts for the Web": Arial, Times New Roman, Courier New, Georgia, Verdana, etc. — genuine MS font files under Microsoft's own redistribution EULA, not metric-compatible substitutes), which is a legitimate, fully-automatable improvement.

It cannot fully satisfy CloakBrowser's check, though: that requires 8 specific fonts (`Segoe UI`, `Segoe UI Light`, `Calibri`, `Marlett`, `MS UI Gothic`, `Franklin Gothic`, `Consolas`, `Courier New`), and 7 of those 8 are Windows-OS/Office-exclusive system fonts with no legal, freely-redistributable source — they're not in `ttf-mscorefonts-installer` or any other apt package. If you have legal access to them (e.g. a licensed Windows/Office install), drop the `.ttf`/`.otf` files into `fonts/` before building — see `fonts/README.md`. The warning is suppressed by default (`CLOAKBROWSER_SUPPRESS_FONT_WARNING=1` in the `Dockerfile`'s `ENV`) since it can otherwise never fully resolve here, and its "shown once" marker lives on the same read-only-at-runtime path as the welcome banner (see the `Dockerfile` comment), so it would otherwise reprint on every single cold start.

## A note on AWS Lambda Layers (the actual `Layers` API field)

The native `Layers` field on `CreateFunction`/`UpdateFunctionConfiguration`
only applies to **zip-package** Lambda functions. Every function in this repo
is `PackageType: Image` (needed because the CloakBrowser Chromium binary and
its dependencies are far too large — and too native/OS-specific — for a zip
layer), and container-image Lambda functions **cannot** attach a `Layers`
resource at all; everything must be baked into the image itself. A shared
**base image** published to ECR is the standard equivalent for container
image functions, and is what this directory provides.

## Build & push

Do this before (re)building any handler image — handler Dockerfiles default
to pulling this image by tag from ECR.

```bash
aws ecr get-login-password --region us-east-1 \
  | docker login --username AWS --password-stdin <account>.dkr.ecr.us-east-1.amazonaws.com

docker buildx build --platform linux/arm64 --provenance=false --sbom=false \
  -t <account>.dkr.ecr.us-east-1.amazonaws.com/bytelyon-cloakbrowser-base:latest \
  --push .
```

`--provenance=false --sbom=false` is required — without it, buildx produces a
manifest-list/attestation image that Lambda's `CreateFunction`/
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
  -t bytelyon-cloakbrowser-base:arm64 --load .
```

Then point a handler at it without touching ECR:

```bash
docker buildx build --platform linux/arm64 --provenance=false --sbom=false \
  --build-arg BASE_IMAGE=bytelyon-cloakbrowser-base:arm64 \
  -t bytelyon-page-scraper:arm64 --load ../page-scraper
```

## Versioning

Prefer pushing a new tag (e.g. `:v2`) over overwriting `:latest` in place,
then bump the `BASE_IMAGE` default `ARG` in every handler `Dockerfile` you
want to pick up the change. Already-deployed Lambda functions are pinned to
the image **digest** Lambda resolved at the most recent
`update-function-code` call — moving `:latest` in ECR doesn't affect them
until each is explicitly redeployed, but two handlers built against `:latest`
at different times can silently end up on different digests. Tagging
deliberately avoids that ambiguity.
