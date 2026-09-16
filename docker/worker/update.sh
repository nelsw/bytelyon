#!/bin/sh
# Rebuilds every image this worker fleet's Python code (common.py, worker.py,
# handlers/*.py) ends up in, in one shot, whenever any of that code changes.
# There are two separate tags this repo actually uses day to day, and it's
# easy to update one and forget the other:
#   - bytelyon-worker:arm64  standalone tag from build.sh's direct buildx
#                            build, for a plain `docker run` per
#                            INSTRUCTIONS.md's "Running a worker" section.
#   - worker-1.0/app:latest  the tag ../../compose.yml's `worker` service
#                            actually runs, built via `docker compose build`
#                            so Sail/`docker compose up worker` picks up the
#                            same change.
#
# Both build FROM the same private ECR base images (bytelyon-cloakbrowser-base,
# bytelyon-seleniumbase-base), so this also refreshes the ECR auth token
# first -- that expires periodically, and buildx otherwise fails with a bare
# "403 Forbidden" / "authorization token has expired" partway through.
#
# Usage: ./update.sh   (run from anywhere -- paths below are script-relative)
set -e

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
repo_root=$(CDPATH= cd -- "$script_dir/../.." && pwd)

aws_region="${AWS_DEFAULT_REGION:-us-east-1}"
aws_account="${AWS_ACCOUNT_ID:-138305277395}"
ecr_host="$aws_account.dkr.ecr.$aws_region.amazonaws.com"

echo "==> Refreshing ECR auth for $ecr_host"
aws ecr get-login-password --region "$aws_region" \
    | docker login --username AWS --password-stdin "$ecr_host"

echo "==> Building bytelyon-worker:arm64 (docker/worker/build.sh)"
(cd "$script_dir" && docker buildx build --platform linux/arm64 --provenance=false --sbom=false -t bytelyon-worker:arm64 --load .)

echo "==> Building worker-1.0/app:latest (docker compose build worker)"
(cd "$repo_root" && docker compose build worker)

echo "==> Done. Both image tags now include the latest Python code."
echo "    Neither is picked up by an already-running container -- recreate"
echo "    whichever one you actually run to pick up the change:"
echo "      docker compose up -d --force-recreate worker      # Sail/compose"
echo "      docker stop bytelyon-worker && docker rm bytelyon-worker && docker run ...  # standalone, see INSTRUCTIONS.md"
