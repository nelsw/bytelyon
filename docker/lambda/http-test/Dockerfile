# Minimal, throwaway diagnostic Lambda: plain `requests` HTTP client, no
# Chromium/Playwright/cloakbrowser at all. Built FROM the AWS-managed Lambda
# Python base image (not ../base-image) since it needs none of the CloakBrowser
# Lambda glue \u2014 just requests (+ PySocks for socks5:// proxy support).
#
# Purpose: determine whether authenticated proxy traffic hangs specifically
# when it originates from AWS Lambda, independent of whether the client is a
# full browser or a plain HTTP client. See lambda_handler.py's docstring.
#
# Build:
#   docker buildx build --platform linux/arm64 --provenance=false --sbom=false \
#     -t bytelyon-http-test:arm64 --load .
#
# Local smoke test:
#   docker run --rm -p 9000:8080 bytelyon-http-test:arm64
#   curl -XPOST http://localhost:9000/2015-03-31/functions/function/invocations \
#     -d '{"proxy":"http://user:pass@host:port"}'

FROM public.ecr.aws/lambda/python:3.12

RUN pip install --no-cache-dir requests "requests[socks]"

COPY lambda_handler.py ${LAMBDA_TASK_ROOT}/lambda_handler.py

CMD ["lambda_handler.handler"]
