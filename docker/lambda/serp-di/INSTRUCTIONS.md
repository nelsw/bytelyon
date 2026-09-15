# serp-di: standalone Lambda (reference/one-off testing)

This directory is now **Lambda-mode only**. The production path — one or
more local-machine workers pulling jobs from a shared SQS queue — moved to
`../../worker/` (a single consolidated location for the serp/news/sitemap
worker fleet). See `../../worker/INSTRUCTIONS.md` for that.

`lambda_handler.py` here is kept as a standalone, full-featured Lambda
function (DataImpulse proxy, GeoIP, SERP parsing, S3 upload, all in one
Python module) — useful for one-off testing or comparing Lambda-vs-local-
worker block rates, but no longer what `SearchBotJob` calls in production.

## Build & deploy (unchanged)

```bash
docker buildx build --platform linux/arm64 --provenance=false --sbom=false \
  --build-arg BASE_IMAGE=138305277395.dkr.ecr.us-east-1.amazonaws.com/bytelyon-cloakbrowser-base:latest \
  -t bytelyon-serp-di:arm64 --load .

docker run --rm -p 9000:8080 bytelyon-serp-di:arm64
curl -XPOST http://localhost:9000/2015-03-31/functions/function/invocations \
  -d '{"query":"sailing blocks"}'
```

Deployed as the `bytelyon-serp-di` Lambda function (arm64, 3008MB, 300s
timeout) in `us-east-1`.

## Event schema / response

See `lambda_handler.py`'s own module docstring — same `data`/
`screenshot_key`/`content_key` contract as `../serp-scraper`.
