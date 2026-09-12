# CloakBrowser fast single-page capture (HTML + full-page screenshot) on
# AWS Lambda.
#
# Built FROM the shared bytelyon-cloakbrowser-base image (see
# ../base-image/Dockerfile and ../base-image/INSTRUCTIONS.md), which carries
# every bit of common CloakBrowser + Lambda glue: awslambdaric, the Lambda
# Runtime Interface Emulator, boto3, the dual-mode entrypoint, the non-root
# readability fix, and the runtime ENV. This Dockerfile only adds this
# handler's own module — see lambda_handler.py.
#
# Sibling of ../news-scraper, ../page-scraper, ../scraper, ../serp-scraper,
# and ../site-crawler (same base image, different lambda_handler.py). This
# handler takes a single URL (with an optional proxy), uploads the rendered
# HTML and a full-page screenshot to S3, and returns just the two S3 keys —
# deliberately minimal, optimized for speed over ../page-scraper's link/meta
# extraction.
#
# ─── Build (the base image must already exist, locally or in ECR) ────────────
#   docker buildx build --platform linux/arm64 --provenance=false --sbom=false \
#     -t bytelyon-grab:arm64 --load .
#
#   docker run --rm -p 9000:8080 \
#     -e AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY -e AWS_SESSION_TOKEN -e AWS_REGION=us-east-1 \
#     bytelyon-grab:arm64
#   curl -XPOST http://localhost:9000/2015-03-31/functions/function/invocations \
#     -d '{"url":"https://example.com"}'
#
# --provenance=false --sbom=false is required: without it, buildx emits a
# manifest-list/attestation image that Lambda's CreateFunction/UpdateFunctionCode
# rejects with "image manifest, config or layer media type ... is not supported".
#
# To build against a locally-loaded (not-yet-pushed) base image instead of
# pulling from ECR, see ../base-image/INSTRUCTIONS.md:
#   --build-arg BASE_IMAGE=bytelyon-cloakbrowser-base:arm64

ARG BASE_IMAGE=138305277395.dkr.ecr.us-east-1.amazonaws.com/bytelyon-cloakbrowser-base:latest
FROM ${BASE_IMAGE}

# Handler sits at /app (already on Python's import path, WORKDIR=/app),
# imports cloakbrowser as a normal library. ENTRYPOINT/CMD are inherited from
# the base image (both already correct for this handler's module name).
COPY lambda_handler.py /app/lambda_handler.py
