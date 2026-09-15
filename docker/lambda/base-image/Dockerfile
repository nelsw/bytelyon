# Shared AWS Lambda base image for every bytelyon-handlers CloakBrowser
# handler — derived from the official CloakHQ image.
#
# `FROM cloakhq/cloakbrowser:<tag>` is an official distribution channel under
# the CloakBrowser Binary License — pulling it isn't redistribution. This
# image layers the Lambda glue that used to be copy-pasted verbatim into
# every sibling handler's own Dockerfile (../news-scraper,
# ../page-scraper, ../scraper, ../serp-scraper, ../site-crawler): the Lambda
# Runtime Interface Client (awslambdaric), the Lambda Runtime Interface
# Emulator (for local `docker run` testing), boto3 (some handlers upload
# screenshots to S3), the dual-mode entrypoint, the non-root readability fix,
# and the Lambda runtime ENV.
#
# Building this once and publishing it to ECR means every handler's own
# Dockerfile shrinks to just:
#   ARG BASE_IMAGE=<this image's ECR URI>
#   FROM ${BASE_IMAGE}
#   COPY lambda_handler.py /app/lambda_handler.py
#
# That's the point of this directory: smaller/faster handler builds and
# pushes (Docker only needs to build+push the thin per-handler layer — the
# shared layers are already in ECR and get reused via layer caching), a
# single place to patch Lambda glue/CVEs instead of five, and no more
# risk of the sibling Dockerfiles quietly drifting apart.
#
# ─── Build & push this image FIRST, before building any handler ──────────────
#   docker buildx build --platform linux/arm64 --provenance=false --sbom=false \
#     -t <account>.dkr.ecr.<region>.amazonaws.com/bytelyon-cloakbrowser-base:latest \
#     --push .
#
# (Use --load instead of --push, and tag it locally, e.g.
# bytelyon-cloakbrowser-base:arm64, to test handler builds against a
# not-yet-published base image via each handler's `BASE_IMAGE` build arg.)
#
# --provenance=false --sbom=false is required here AND in every handler image
# built FROM this one: without it, buildx emits a manifest-list/attestation
# image that Lambda's CreateFunction/UpdateFunctionCode rejects with
# "image manifest, config or layer media type ... is not supported".
#
# Pin a specific tag (e.g. cloakhq/cloakbrowser:0.3.25), via
# `--build-arg CLOAKBROWSER_TAG=0.3.25`, for reproducible builds; `latest`
# floats with CloakHQ's release cadence. When you bump it, push this image
# under a new tag (e.g. :v2) rather than overwriting :latest in place, then
# bump the `BASE_IMAGE` default in every handler Dockerfile to match — that
# keeps already-deployed Lambda functions pinned to the digest they were last
# updated with (Lambda resolves and freezes the digest at
# create/update-function-code time) until each is explicitly redeployed.

ARG CLOAKBROWSER_TAG=latest
FROM cloakhq/cloakbrowser:${CLOAKBROWSER_TAG}

# ─── Lambda Runtime Interface Client + S3 client ──────────────────────────────
# Neither is bundled in the canonical image (it's not the AWS-managed Lambda
# Python base image). boto3 is only imported by the handlers that upload
# screenshots to S3 (page-scraper, site-crawler) but is baked in for every
# handler here anyway — it's a small, pure-Python dependency, and shipping one
# base image that works for all handlers is the entire point of this file.
RUN pip install --no-cache-dir awslambdaric boto3

# ─── Lambda Runtime Interface Emulator (local `docker run` testing) ───────────
# Bundled into the image so users can hit the standard local-invoke endpoint
# without mounting the RIE separately. TARGETARCH is provided by buildx.
ARG TARGETARCH
ADD https://github.com/aws/aws-lambda-runtime-interface-emulator/releases/latest/download/aws-lambda-rie-${TARGETARCH} \
    /usr/local/bin/aws-lambda-rie
RUN chmod +x /usr/local/bin/aws-lambda-rie

# ─── Lambda glue ──────────────────────────────────────────────────────────────
# Dual-mode entrypoint replaces the canonical bin/docker-entrypoint.sh: same
# Xvfb startup, plus routing for `module.func` CMDs through awslambdaric. Xvfb
# isn't strictly required by any current handler (they all launch
# headless=True by default), but it's kept so the other invocation surfaces
# below (python, cloakserve, etc.) still support headed mode exactly like the
# canonical image.
COPY lambda-entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# ─── Baked-in GeoIP database (for cloakbrowser's optional geoip=True) ─────────
# cloakbrowser's geoip.py downloads GeoLite2-City.mmdb (~65 MB) on first use
# into get_cache_dir()/geoip/, which — because CLOAKBROWSER_CACHE_DIR (set
# below) points at /root/.cloakbrowser, the same read-only-at-runtime path as
# the baked Chromium binary — fails silently in real Lambda (mkdir on a
# read-only FS), surfacing as "GeoIP resolution failed: GeoIP database is
# unavailable" from any handler that passes geoip=True. Baking the DB in at
# build time avoids any runtime write entirely: _ensure_geoip_db() just finds
# the file already present and returns immediately. Same P3TERX MaxMind
# GeoLite2-City mirror cloakbrowser itself downloads from (no license key
# needed); re-bake periodically by rebuilding this image to pick up updates
# (cloakbrowser's own 30-day background refresh can't write here either, but
# fails silently in a background thread without affecting resolution).
ADD https://github.com/P3TERX/GeoLite.mmdb/raw/download/GeoLite2-City.mmdb \
    /root/.cloakbrowser/geoip/GeoLite2-City.mmdb

# ─── Real Microsoft fonts (for Windows-fingerprint spoofing) ─────────────────
# CloakBrowser's own guidance (https://github.com/CloakHQ/cloakbrowser#font-setup-on-linux):
# a bare Linux image's font list is itself a detection signal when spoofing a
# Windows fingerprint — every launch otherwise logs "Incomplete Windows font
# set". `ttf-mscorefonts-installer` is Debian's packaging of Microsoft's own
# "Core Fonts for the Web" (Arial, Times New Roman, Courier New, Georgia,
# Verdana, ...) — genuine Microsoft font files fetched from Sourceforge at
# build time under Microsoft's redistribution EULA (auto-accepted below via
# debconf-set-selections, the standard hands-off way to script this package;
# lives in the "contrib" component, which isn't enabled by default). This is
# real Microsoft font data, not a metric-compatible substitute like
# fonts-liberation (already present upstream, and left in place — more
# font coverage only helps).
#
# CloakBrowser's own docs call this "not enough" by itself — a real Windows
# install has 200+ fonts, Office adds ~50 more, versus ~13 here. See
# ./fonts/README.md for how to add the rest if you have legal access to them;
# that directory is empty by default (just its own README, which fc-cache
# silently ignores) so this COPY is always valid even with nothing added.
RUN sed -i 's/Components: main$/Components: main contrib/' /etc/apt/sources.list.d/debian.sources \
    && echo "ttf-mscorefonts-installer msttcorefonts/accepted-mscorefonts-eula select true" | debconf-set-selections \
    && apt-get update \
    && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends ttf-mscorefonts-installer fontconfig \
    && rm -rf /var/lib/apt/lists/*
COPY fonts/ /usr/share/fonts/windows/
RUN fc-cache -f

# ─── Lambda non-root readability fix ──────────────────────────────────────────
# The canonical image bakes the Chromium binary at /root/.cloakbrowser/ (root's
# HOME at build time). Lambda runs the container as a non-root user that can't
# read /root by default (mode 750). Make the whole binary tree world-readable
# and traversable.
RUN chmod -R o+rX /root /root/.cloakbrowser

# ─── Silence the CloakBrowser free-tier welcome banner ────────────────────────
# Every keyless launch, cloakbrowser.download._show_welcome() prints a
# multi-line "CloakBrowser - stealth Chromium for automation ... Star us if
# CloakBrowser helps your project!" banner to stderr (CloudWatch, for a Lambda
# handler), gated by a cache-dir marker file (.welcome_shown): shown once if
# missing, and re-shown on the free tier once the marker is older than 24h
# (WELCOME_FREE_INTERVAL). Baking in a fresh marker at build time (as this
# image used to) only defers the banner by a day — a long-lived ECR image
# would start reprinting it on cold starts once the marker goes stale, and
# Lambda's read-only runtime FS means nothing can ever refresh it. Patch the
# one line that gates the banner so _show_welcome() always returns
# immediately, regardless of tier or marker state. Fails the build loudly
# (assert) if a future upstream release reshapes this function, rather than
# silently leaving the banner in place.
RUN python3 -c 'import pathlib, cloakbrowser.download as m; p = pathlib.Path(m.__file__); src = p.read_text(); anchor = "    marker = get_cache_dir() / \".welcome_shown\"\n    if not _welcome_due(marker, pro=(tier == \"pro\")):\n        return\n"; assert anchor in src, "cloakbrowser _show_welcome anchor not found - upstream source changed"; p.write_text(src.replace(anchor, "    return  # bytelyon-cloakbrowser-base: banner disabled (see base-image/Dockerfile)\n" + anchor, 1))'

# ─── Lambda runtime env ───────────────────────────────────────────────────────
# HOME=/tmp gives Chromium a writable scratch dir (Lambda only allows writes
# under /tmp). CLOAKBROWSER_CACHE_DIR points at the baked binary location since
# HOME=/tmp would otherwise make get_cache_dir() resolve to /tmp/.cloakbrowser
# (empty). Auto-update is disabled because the runtime FS is read-only.
#
# CLOAKBROWSER_SUPPRESS_FONT_WARNING: same read-only-FS problem as the welcome
# banner above, applied to a check we can only ever partially satisfy. The
# "Incomplete Windows font set" warning (browser.py's _maybe_warn_windows_fonts)
# requires ALL of 8 specific fonts to be present: Segoe UI, Segoe UI Light,
# Calibri, Marlett, MS UI Gothic, Franklin Gothic, Consolas, Courier New. The
# real-Microsoft-fonts step above (ttf-mscorefonts-installer) only covers
# Courier New — the other 7 are Windows-OS/Office-exclusive system fonts with
# no legal, freely-redistributable source; they aren't part of any apt package,
# core-fonts-for-the-web included (see ./fonts/README.md if you have licensed
# access to add them yourself). Since the check can never fully pass here, and
# its "shown once" marker lives under this same read-only CLOAKBROWSER_CACHE_DIR
# (the write silently fails every time, so without this it would otherwise
# reprint on every single cold start, forever — confirmed in practice),
# suppress it outright rather than let an unactionable warning spam CloudWatch.
ENV HOME=/tmp \
    CLOAKBROWSER_CACHE_DIR=/root/.cloakbrowser \
    CLOAKBROWSER_AUTO_UPDATE=false \
    CLOAKBROWSER_SUPPRESS_FONT_WARNING=1

ENTRYPOINT ["/entrypoint.sh"]

# Every handler ships a module literally named lambda_handler.py with a
# `handler(event, context)` function (see each sibling's own Dockerfile,
# which COPYs its handler to /app/lambda_handler.py), so this default CMD is
# already correct for all of them and none needs to repeat it. Override it in
# a handler's own Dockerfile only if that convention ever changes.
CMD ["lambda_handler.handler"]

# ─── Same as the canonical CloakHQ image (CMD overridden per-handler) ─────────
#   docker run --rm -it <handler-image>:arm64 python                          # REPL
#   docker run --rm <handler-image>:arm64 python examples/basic.py            # examples
#   docker run --rm -p 9222:9222 <handler-image>:arm64 cloakserve --port=9222 # CDP server
#   docker run --rm <handler-image>:arm64 cloaktest                           # stealth tests
#   docker run --rm -it <handler-image>:arm64 node                            # JS wrapper
#   docker run --rm -it <handler-image>:arm64 bash                           # shell
