# SeleniumBase (Stealthy Playwright Mode) fast single-page capture (HTML +
# full-page screenshot) on AWS Lambda.
#
# Built FROM the shared bytelyon-seleniumbase-base image (see
# ../../images/seleniumbase-base/Dockerfile and
# ../../images/seleniumbase-base/INSTRUCTIONS.md), which carries every bit of
# common Lambda glue (awslambdaric, the Lambda Runtime Interface Emulator,
# boto3, the dual-mode entrypoint) plus system Chromium, `seleniumbase`, and
# `playwright`, and the `seleniumbase_playwright.launch()` shim that connects
# a normal Playwright `Browser` to SeleniumBase's stealthy CDP Mode Chromium.
# This Dockerfile only adds this handler's own module — see
# lambda_handler.py.
#
# This is the SeleniumBase counterpart of ../grab: same event schema, same
# S3 upload behavior, same response shape. The only difference is the
# stealth-browser provider — SeleniumBase's CDP Mode
# (https://seleniumbase.io/) instead of CloakBrowser — see
# ../../images/seleniumbase-base/seleniumbase_playwright.py for how that
# swap is bridged behind an (almost) identical `launch()` call.
#
# ─── Build (the base image must already exist, locally or in ECR) ────────────
#   docker buildx build --platform linux/arm64 --provenance=false --sbom=false \
#     -t bytelyon-grab-seleniumbase:arm64 --load .
#
#   docker run --rm -p 9000:8080 \
#     -e AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY -e AWS_SESSION_TOKEN -e AWS_REGION=us-east-1 \
#     bytelyon-grab-seleniumbase:arm64
#   curl -XPOST http://localhost:9000/2015-03-31/functions/function/invocations \
#     -d '{"url":"https://example.com"}'
#
# --provenance=false --sbom=false is required: without it, buildx emits a
# manifest-list/attestation image that Lambda's CreateFunction/UpdateFunctionCode
# rejects with "image manifest, config or layer media type ... is not supported".
#
# To build against a locally-loaded (not-yet-pushed) base image instead of
# pulling from ECR, see ../../images/seleniumbase-base/INSTRUCTIONS.md:
#   --build-arg BASE_IMAGE=bytelyon-seleniumbase-base:arm64

ARG BASE_IMAGE=138305277395.dkr.ecr.us-east-1.amazonaws.com/bytelyon-seleniumbase-base:latest
FROM ${BASE_IMAGE}

# Handler sits at /app (already on Python's import path, WORKDIR=/app),
# imports playwright.sync_api types plus this repo's seleniumbase_playwright
# shim (on PYTHONPATH via /opt/python). ENTRYPOINT/CMD are inherited from the
# base image (both already correct for this handler's module name).
COPY lambda_handler.py /app/lambda_handler.py
