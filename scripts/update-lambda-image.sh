#!/usr/bin/env bash
#
# Rebuilds a handler's Docker image and updates the matching AWS Lambda
# function's code to the freshly pushed image — the same manual steps
# documented in every handler's own INSTRUCTIONS.md, wrapped into one
# command.
#
# Usage:
#   scripts/update-lambda-image.sh <function-name> [options]
#
# What it does:
#   1. Looks up the function's current Code.ImageUri and Architectures[0]
#      via `aws lambda get-function`/`get-function-configuration`.
#   2. Derives the ECR repository + tag to (re)build from that ImageUri —
#      reuses whatever tag is already deployed, so this script doesn't
#      silently introduce a brand-new tag; falls back to "latest" if the
#      function is currently pinned to a bare digest with no tag.
#   3. `docker buildx build`s the handler directory (--provenance=false
#      --sbom=false, required for Lambda — see any handler's Dockerfile),
#      tagged to match step 2, `--load`ed into the local image store.
#   4. `docker push`es it and captures the resulting digest.
#   5. `aws lambda update-function-code` with `--image-uri <repo>@<digest>`,
#      pinning the function to that exact digest (not a floating tag).
#   6. Waits for the update to finish and prints the resulting state.
#
# Requires: docker (with buildx), aws CLI v2, already authenticated with
# sufficient IAM permissions for ECR (push) and Lambda
# (GetFunction/GetFunctionConfiguration/UpdateFunctionCode) in the target
# account/region.
#
# The handler directory is assumed to be named after the function name with
# the "bytelyon-" prefix stripped (e.g. "bytelyon-news-scraper" ->
# "news-scraper"), matching every handler in this repo. Pass --dir to
# override this for a function that doesn't follow that convention.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

usage() {
  cat <<'EOF'
Usage: scripts/update-lambda-image.sh <function-name> [options]

Options:
  --region REGION    AWS region (default: $AWS_REGION, $AWS_DEFAULT_REGION,
                      else "us-east-1")
  --dir PATH         Handler directory to build (default: <repo-root>/<function
                      name with the "bytelyon-" prefix stripped>)
  --platform ARCH    Docker platform architecture to build for, e.g. "arm64"
                      or "amd64" (default: the function's own configured
                      Architectures[0], so it matches what's already deployed)
  --base-image URI   Override the BASE_IMAGE build-arg (default: whatever the
                      handler's own Dockerfile defaults to)
  --dry-run          Print what would happen without building, pushing, or
                      updating anything
  -h, --help         Show this help

Examples:
  scripts/update-lambda-image.sh bytelyon-news-scraper
  scripts/update-lambda-image.sh bytelyon-warmer --dry-run
  scripts/update-lambda-image.sh bytelyon-serp-scraper --dir serp-scraper --region us-east-1
EOF
}

FUNCTION_NAME=""
REGION="${AWS_REGION:-${AWS_DEFAULT_REGION:-us-east-1}}"
HANDLER_DIR=""
PLATFORM_ARCH="arm64"
BASE_IMAGE_OVERRIDE=""
DRY_RUN=false

while [ $# -gt 0 ]; do
  case "$1" in
    --region)
      REGION="$2"; shift 2 ;;
    --dir)
      HANDLER_DIR="$2"; shift 2 ;;
    --platform)
      PLATFORM_ARCH="$2"; shift 2 ;;
    --base-image)
      BASE_IMAGE_OVERRIDE="$2"; shift 2 ;;
    --dry-run)
      DRY_RUN=true; shift ;;
    -h|--help)
      usage; exit 0 ;;
    -*)
      echo "Unknown option: $1" >&2; usage; exit 1 ;;
    *)
      if [ -z "$FUNCTION_NAME" ]; then
        FUNCTION_NAME="$1"
      else
        echo "Unexpected extra argument: $1" >&2; exit 1
      fi
      shift
      ;;
  esac
done

if [ -z "$FUNCTION_NAME" ]; then
  echo "Error: function name is required." >&2
  usage
  exit 1
fi

if [ -z "$HANDLER_DIR" ]; then
  DEFAULT_DIR_NAME="${FUNCTION_NAME#bytelyon-}"
  HANDLER_DIR="${REPO_ROOT}/${DEFAULT_DIR_NAME}"
fi

if [ ! -f "${HANDLER_DIR}/Dockerfile" ]; then
  echo "Error: no Dockerfile found at '${HANDLER_DIR}'." >&2
  echo "Pass --dir to point at the correct handler directory (e.g. --dir ${REPO_ROOT}/serp-scraper)." >&2
  exit 1
fi

echo "==> Looking up function '${FUNCTION_NAME}' in region '${REGION}'..."
CURRENT_IMAGE_URI="$(aws lambda get-function \
  --function-name "$FUNCTION_NAME" --region "$REGION" \
  --query "Code.ImageUri" --output text)"
CURRENT_ARCH="$(aws lambda get-function-configuration \
  --function-name "$FUNCTION_NAME" --region "$REGION" \
  --query "Architectures[0]" --output text)"

if [ -z "$PLATFORM_ARCH" ]; then
  PLATFORM_ARCH="$CURRENT_ARCH"
fi

# Derive the ECR repo + tag to build/push from the function's *current*
# ImageUri: reuse whatever tag is already deployed rather than inventing a
# new one, falling back to "latest" only if the function is currently
# pinned to a bare digest (no tag at all) — e.g. right after this same
# script last ran.
if [[ "$CURRENT_IMAGE_URI" == *"@sha256:"* ]]; then
  REPO="${CURRENT_IMAGE_URI%%@*}"
  TAG="latest"
elif [[ "$CURRENT_IMAGE_URI" == *:* ]]; then
  REPO="${CURRENT_IMAGE_URI%:*}"
  TAG="${CURRENT_IMAGE_URI##*:}"
else
  REPO="$CURRENT_IMAGE_URI"
  TAG="latest"
fi

REGISTRY_HOST="${REPO%%/*}"

echo "==> Function:       ${FUNCTION_NAME}"
echo "==> Handler dir:    ${HANDLER_DIR}"
echo "==> Current image:  ${CURRENT_IMAGE_URI}"
echo "==> Target repo:tag ${REPO}:${TAG}"
echo "==> Platform:       linux/${PLATFORM_ARCH}"

if $DRY_RUN; then
  echo "==> --dry-run: stopping before build/push/update."
  exit 0
fi

echo "==> Logging in to ECR (${REGISTRY_HOST})..."
aws ecr get-login-password --region "$REGION" \
  | docker login --username AWS --password-stdin "$REGISTRY_HOST"

BUILD_ARGS=()
if [ -n "$BASE_IMAGE_OVERRIDE" ]; then
  BUILD_ARGS+=(--build-arg "BASE_IMAGE=${BASE_IMAGE_OVERRIDE}")
fi

echo "==> Building ${REPO}:${TAG} from ${HANDLER_DIR} (linux/${PLATFORM_ARCH})..."
docker buildx build \
  --platform "linux/${PLATFORM_ARCH}" \
  --provenance=false --sbom=false \
  "${BUILD_ARGS[@]+"${BUILD_ARGS[@]}"}" \
  -t "${REPO}:${TAG}" \
  --load \
  "$HANDLER_DIR"

echo "==> Pushing ${REPO}:${TAG}..."
PUSH_OUTPUT="$(docker push "${REPO}:${TAG}")"
echo "$PUSH_OUTPUT"

DIGEST="$(echo "$PUSH_OUTPUT" | grep -oE 'sha256:[0-9a-f]{64}' | tail -1)"
if [ -z "$DIGEST" ]; then
  echo "Error: couldn't determine the pushed image digest from 'docker push' output." >&2
  exit 1
fi

IMAGE_URI="${REPO}@${DIGEST}"
echo "==> Pushed digest:  ${DIGEST}"

echo "==> Updating function code to ${IMAGE_URI}..."
aws lambda update-function-code \
  --function-name "$FUNCTION_NAME" \
  --image-uri "$IMAGE_URI" \
  --region "$REGION" \
  --query "{State:State,LastUpdateStatus:LastUpdateStatus,CodeSha256:CodeSha256}"

echo "==> Waiting for the update to finish..."
aws lambda wait function-updated-v2 --function-name "$FUNCTION_NAME" --region "$REGION"

echo "==> Done:"
aws lambda get-function --function-name "$FUNCTION_NAME" --region "$REGION" \
  --query "{FunctionName:Configuration.FunctionName,State:Configuration.State,LastUpdateStatus:Configuration.LastUpdateStatus,CodeSha256:Configuration.CodeSha256,ImageUri:Code.ImageUri}"
