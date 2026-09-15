# Shared AWS Lambda base image for every bytelyon-handlers SeleniumBase
# handler — unlike ../../lambda/base-image (which layers Lambda glue on top
# of the official `cloakhq/cloakbrowser` Docker Hub image), there's no
# official pre-built SeleniumBase image to start from, so this is a from
# scratch ("OS-only") Lambda base image: a plain Debian-based Python image,
# system Chromium, and the same Lambda Runtime Interface Client / Emulator /
# entrypoint glue as the CloakBrowser base image.
#
# SeleniumBase (https://seleniumbase.io/) is used here as the *stealth
# browser provider for Playwright* — see seleniumbase_playwright.py, baked
# into this image. SeleniumBase's "Stealthy Playwright Mode" launches its
# hardened/patched Chromium via CDP Mode (`seleniumbase.sb_cdp.Chrome()`),
# then hands a normal `playwright.sync_api` `Browser` back, connected to
# that already-running, already-stealthy browser over the Chrome DevTools
# Protocol (`playwright.chromium.connect_over_cdp(...)`). Every downstream
# handler built FROM this image still writes ordinary Playwright code
# (`page.goto()`, `page.content()`, `page.screenshot()`, ...) — only the
# *launch* call differs from the CloakBrowser-based handlers:
#   from seleniumbase_playwright import launch   # instead of: from cloakbrowser import launch
#
# ─── Build & push this image FIRST, before building any *-seleniumbase
#     handler ──────────────────────────────────────────────────────────────
#   docker buildx build --platform linux/arm64 --provenance=false --sbom=false \
#     -t <account>.dkr.ecr.<region>.amazonaws.com/bytelyon-seleniumbase-base:latest \
#     --push .
#
# (Use --load instead of --push, and tag it locally, e.g.
# bytelyon-seleniumbase-base:arm64, to test handler builds against a
# not-yet-published base image via each handler's `BASE_IMAGE` build arg —
# same pattern as ../../lambda/base-image/INSTRUCTIONS.md.)
#
# --provenance=false --sbom=false is required here AND in every handler image
# built FROM this one: without it, buildx emits a manifest-list/attestation
# image that Lambda's CreateFunction/UpdateFunctionCode rejects with
# "image manifest, config or layer media type ... is not supported".
#
# Push new revisions under a new tag (e.g. :v2) rather than overwriting
# :latest in place, then bump the `BASE_IMAGE` default in every handler
# Dockerfile that should pick it up — same versioning rationale as
# ../../lambda/base-image/INSTRUCTIONS.md.

ARG PYTHON_VERSION=3.12
FROM python:${PYTHON_VERSION}-slim-bookworm

WORKDIR /app

# ─── System Chromium (the actual browser SeleniumBase's CDP Mode drives) ──────
# SeleniumBase's CDP Mode launches the browser directly as a subprocess and
# speaks the Chrome DevTools Protocol to it over its own
# --remote-debugging-port — no chromedriver/webdriver binary is involved at
# all (that's only needed for SeleniumBase's classic Selenium/WebDriver
# mode, which this image does not set up). Debian's `chromium` package is
# used instead of Google Chrome because Google only publishes amd64 Linux
# builds, and this image (like every other image in this repo) targets
# arm64 as well. `find_chrome_executable()`
# (seleniumbase/undetected/cdp_driver/config.py, vendored inside the
# `seleniumbase` pip package) auto-discovers `chromium`/`chromium-browser`
# on PATH, so no `browser_executable_path=` override is needed at call time.
# `xvfb` supports the dual-mode entrypoint's headed invocation surfaces,
# same rationale as ../../lambda/base-image/Dockerfile, even though no
# current handler needs headed mode (all launch with headless=True).
RUN apt-get update && apt-get install -y --no-install-recommends \
        chromium \
        xvfb \
        ca-certificates \
        fonts-liberation \
    && rm -rf /var/lib/apt/lists/*

# ─── Lambda Runtime Interface Client + S3 client + browser automation ─────────
# awslambdaric / boto3: identical rationale to the CloakBrowser base image
# (../../lambda/base-image/Dockerfile) — this is not the AWS-managed Lambda
# Python base image, so neither is bundled by default.
# seleniumbase: provides CDP Mode (`seleniumbase.sb_cdp`), which launches
# and drives the stealthy Chromium instance installed above.
# playwright: the *client* library only (`playwright.sync_api`) — used
# purely to `connect_over_cdp()` to the browser SeleniumBase already
# launched. `playwright install` (which downloads Playwright's own managed
# browser binaries) is deliberately never run: this image always drives the
# system Chromium installed above, never a Playwright-managed one.
RUN pip install --no-cache-dir \
        awslambdaric \
        boto3 \
        seleniumbase \
        playwright

# ─── Lambda Runtime Interface Emulator (local `docker run` testing) ───────────
# Bundled into the image so users can hit the standard local-invoke endpoint
# without mounting the RIE separately. TARGETARCH is provided by buildx.
ARG TARGETARCH
ADD https://github.com/aws/aws-lambda-runtime-interface-emulator/releases/latest/download/aws-lambda-rie-${TARGETARCH} \
    /usr/local/bin/aws-lambda-rie
RUN chmod +x /usr/local/bin/aws-lambda-rie

# ─── Lambda glue ──────────────────────────────────────────────────────────────
# Dual-mode entrypoint: same Xvfb startup + `module.func` CMD routing through
# awslambdaric as ../../lambda/base-image/lambda-entrypoint.sh (verbatim
# copy — the logic is generic, not CloakBrowser-specific).
COPY lambda-entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# ─── bytelyon Playwright-over-SeleniumBase-CDP glue module ────────────────────
# Every handler built FROM this image imports `seleniumbase_playwright.launch()`
# instead of `cloakbrowser.launch()`. Baked in here (rather than duplicated
# per-handler) for the same reason the CloakBrowser base image centralizes
# its Lambda glue: one place to patch, smaller/faster per-handler layers, no
# risk of sibling handlers' copies drifting apart. Installed under /opt/python
# (added to PYTHONPATH below) rather than next to /app/lambda_handler.py so
# it isn't shadowed/duplicated by each handler's own thin COPY layer.
COPY seleniumbase_playwright.py /opt/python/seleniumbase_playwright.py
ENV PYTHONPATH=/opt/python

# ─── Lambda runtime env ───────────────────────────────────────────────────────
# HOME=/tmp gives Chromium (and SeleniumBase's own per-launch temp profile
# dirs) a writable scratch dir — Lambda only allows writes under /tmp.
# Unlike the CloakBrowser base image, no non-root readability fix is needed
# here: the Chromium binary comes from an apt package under /usr/lib and
# /usr/bin (world-readable/executable by normal Debian packaging), not from
# a build-time download into root's HOME.
ENV HOME=/tmp

ENTRYPOINT ["/entrypoint.sh"]

# Every handler ships a module literally named lambda_handler.py with a
# `handler(event, context)` function (see ../../lambda/grab-seleniumbase's
# own Dockerfile, which COPYs its handler to /app/lambda_handler.py), so
# this default CMD is already correct for all of them.
CMD ["lambda_handler.handler"]

# ─── Local invocation / debugging surfaces ─────────────────────────────────────
#   docker run --rm -it <handler-image>:arm64 python                          # REPL
#   docker run --rm -it <handler-image>:arm64 bash                           # shell
