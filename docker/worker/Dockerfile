# Consolidated local-machine SQS worker for every bytelyon scrape job type
# (serp, news, sitemap, ...) — one image, one queue, one location. See
# worker.py's module docstring for the full architecture and env-var
# contract, and handlers/*.py for the per-job-type navigation logic.
#
# Built FROM the shared bytelyon-cloakbrowser-base image (see
# ../lambda/base-image/Dockerfile) — same GeoIP database, real MS fonts,
# Xvfb entrypoint, and non-root readability fix every Lambda handler in
# ../lambda/ already gets. boto3 and requests are already present there;
# nothing extra needed for handlers/serp.py.
#
# ALSO layers in SeleniumBase's Stealthy Playwright Mode (Debian's system
# `chromium` + the `seleniumbase` pip package + the
# ../images/seleniumbase-base/seleniumbase_playwright.py launch shim,
# copied straight out of that already-published image rather than
# duplicated here — see the COPY --from= below) as a *second* browser
# provider for handlers/generic.py (news/sitemap). Why bother running two
# browser stacks in one image: CloakBrowser Pro's license caps concurrent
# sessions at 5 seats. handlers/serp.py actually needs CloakBrowser's
# Google-specific stealth patches and reserves a seat for that; news and
# sitemap targets are far less aggressive about bot detection and don't —
# SeleniumBase's CDP Mode is plenty for them, and defaulting news/sitemap to
# it (see handlers/generic.py's GENERIC_BROWSER_PROVIDER) means those jobs
# never compete with serp jobs for one of the 5 coveted cloak seats, no
# matter how many run concurrently.
#
# This never gets deployed to AWS Lambda — it's meant to run as a long-lived
# container on a home/office network machine, so there's no Lambda-specific
# build flag concern here (no --provenance=false --sbom=false needed, but
# harmless to keep for consistency with every other image in this repo).
#
# ─── Build ─────────────────────────────────────────────────────────────────
#   docker buildx build --platform linux/arm64 --provenance=false --sbom=false \
#     -t bytelyon-worker:arm64 --load .
#
# (Swap --platform for your machine's actual architecture — build natively
# for whatever's running the container, there's no reason to cross-build
# something that never leaves this machine.)
#
# ─── Run (see worker.py's docstring for what each env var does) ────────────
#   docker run -d --name bytelyon-worker --restart unless-stopped \
#     -e AWS_ACCESS_KEY_ID=... \
#     -e AWS_SECRET_ACCESS_KEY=... \
#     -e AWS_DEFAULT_REGION=us-east-1 \
#     -e SCRAPE_JOBS_QUEUE_URL=https://sqs.us-east-1.amazonaws.com/<account>/bytelyon-scrape-jobs \
#     -e LARAVEL_BASE_URL=http://host.docker.internal \
#     -e LARAVEL_WORKER_TOKEN=... \
#     bytelyon-worker:arm64
#
#   docker logs -f bytelyon-worker

ARG SELENIUMBASE_BASE_IMAGE=138305277395.dkr.ecr.us-east-1.amazonaws.com/bytelyon-seleniumbase-base:latest
ARG BASE_IMAGE=138305277395.dkr.ecr.us-east-1.amazonaws.com/bytelyon-cloakbrowser-base:latest

FROM ${SELENIUMBASE_BASE_IMAGE} AS seleniumbase_source

FROM ${BASE_IMAGE}

# Debian's `chromium` package — the actual browser SeleniumBase's CDP Mode
# drives (see ../images/seleniumbase-base/Dockerfile for the full rationale,
# identical here). `xvfb`/`ca-certificates`/`fonts-liberation` are already
# present via the CloakBrowser base image, so only `chromium` itself is new.
RUN apt-get update && apt-get install -y --no-install-recommends \
        chromium \
    && rm -rf /var/lib/apt/lists/*

# seleniumbase + playwright's *client* library. playwright is already
# present (CloakBrowser depends on it too), but pip will no-op rather than
# error if versions already satisfy seleniumbase's requirement.
RUN pip install --no-cache-dir seleniumbase

# The launch() shim itself — single source of truth stays
# ../images/seleniumbase-base/seleniumbase_playwright.py; copied out of that
# already-published image instead of duplicating the file here so the two
# never drift apart. Same /opt/python + PYTHONPATH placement as that image,
# so it isn't shadowed by this Dockerfile's own /app COPY layer below.
COPY --from=seleniumbase_source /opt/python/seleniumbase_playwright.py /opt/python/seleniumbase_playwright.py
ENV PYTHONPATH=/opt/python

COPY common.py worker.py /app/
COPY handlers/ /app/handlers/

# Override the base image's Lambda-shaped CMD — this container is never
# invoked by the Lambda Runtime API, it just runs worker.py's poll loop
# directly under Xvfb (started by the inherited /entrypoint.sh, same as
# every headed launch in this image family).
ENTRYPOINT ["/entrypoint.sh"]
CMD ["python3", "worker.py"]
