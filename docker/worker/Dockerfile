# Consolidated local-machine SQS worker for every bytelyon scrape job type
# (serp, news, sitemap, ...) — one image, one queue, one location. See
# worker.py's module docstring for the full architecture and env-var
# contract, and handlers/*.py for the per-job-type navigation logic.
#
# Built FROM the shared bytelyon-cloakbrowser-base image (see
# ../lambda/base-image/Dockerfile) — same GeoIP database, real MS fonts,
# Xvfb entrypoint, and non-root readability fix every Lambda handler in
# ../lambda/ already gets. boto3 and requests are already present there;
# nothing extra needed here.
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

ARG BASE_IMAGE=138305277395.dkr.ecr.us-east-1.amazonaws.com/bytelyon-cloakbrowser-base:latest
FROM ${BASE_IMAGE}

COPY common.py worker.py /app/
COPY handlers/ /app/handlers/

# Override the base image's Lambda-shaped CMD — this container is never
# invoked by the Lambda Runtime API, it just runs worker.py's poll loop
# directly under Xvfb (started by the inherited /entrypoint.sh, same as
# every headed launch in this image family).
ENTRYPOINT ["/entrypoint.sh"]
CMD ["python3", "worker.py"]
