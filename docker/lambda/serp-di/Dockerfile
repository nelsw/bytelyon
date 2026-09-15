# CloakBrowser Google SERP scraper on AWS Lambda — DataImpulse SOCKS5 proxy,
# wired up exactly the way it ran successfully overnight as a standalone
# `uv run` script (see lambda_handler.py's module docstring for the original).
#
# Built FROM the shared bytelyon-cloakbrowser-base image (see
# ../base-image/Dockerfile), same as ../serp-scraper — gets the baked-in
# GeoIP database, real MS fonts, Xvfb entrypoint, and non-root readability fix
# for free. A sibling of ../serp-scraper: same retry-on-block/relaunch
# philosophy and the same structured `data` SERP-parsing output contract
# (ported verbatim, see lambda_handler.py), but reaches Google through
# DataImpulse's SOCKS5 proxy instead of Oxylabs, and its own
# launch()/goto()/type() navigation sequence is the one already proven
# reliable overnight as a plain script.
#
# This same image (unmodified) is also the base for ../serp-di/Dockerfile.worker
# — a local-machine SQS worker that imports lambda_handler.py's `main()`
# directly and runs the identical browser logic outside Lambda entirely, since
# Google trusts residential/home-network egress far more than any datacenter
# IP (Lambda's own included). See worker.py's module docstring.
#
# ─── Build (the base image must already exist, locally or in ECR) ────────────
#   docker buildx build --platform linux/arm64 --provenance=false --sbom=false \
#     -t bytelyon-serp-di:arm64 --load .
#
#   docker run --rm -p 9000:8080 bytelyon-serp-di:arm64
#   curl -XPOST http://localhost:9000/2015-03-31/functions/function/invocations \
#     -d '{"query":"sailing blocks"}'
#
# --provenance=false --sbom=false is required: without it, buildx emits a
# manifest-list/attestation image that Lambda's CreateFunction/UpdateFunctionCode
# rejects with "image manifest, config or layer media type ... is not supported".
#
# To build against a locally-loaded (not-yet-pushed) base image instead of
# pulling from ECR:
#   --build-arg BASE_IMAGE=bytelyon-cloakbrowser-base:arm64

ARG BASE_IMAGE=138305277395.dkr.ecr.us-east-1.amazonaws.com/bytelyon-cloakbrowser-base:latest
FROM ${BASE_IMAGE}

# `requests` + PySocks (`requests[socks]`), used by lambda_handler.py's
# `_resolve_geo()` instead of cloakbrowser's own geoip=True. Confirmed via the
# bytelyon-http-test diagnostic Lambda: this same SOCKS5 proxy responds to
# `requests` from AWS Lambda's network in under a second, every time, while
# cloakbrowser's httpx-based geoip resolution reliably hangs to its full
# timeout when run inside Lambda specifically (not reproducible locally in
# Docker, and not reproducible with a different HTTP client from the same
# Lambda function) — an httpx/socksio incompatibility with Lambda's sandboxed
# network namespace, not a proxy problem. geoip2 (for reading the DB baked
# into the base image) and cloakbrowser's own country->locale map are already
# available via the `cloakbrowser[geoip]` extra baked into this base image.
RUN pip install --no-cache-dir requests "requests[socks]"

# Handler sits at /app (already on Python's import path, WORKDIR=/app),
# imports cloakbrowser as a normal library. ENTRYPOINT/CMD are inherited from
# the base image (both already correct for this handler's module name).
COPY lambda_handler.py /app/lambda_handler.py
