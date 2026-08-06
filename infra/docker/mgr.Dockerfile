FROM golang:alpine AS builder

LABEL maintainer="Connor Van Elswyk"

ENV GO111MODULE=on \
    CGO_ENABLED=0

WORKDIR /build

COPY ./apps/mgr .

RUN go mod tidy
RUN go build --ldflags "-s -w -extldflags -static" -o main .

FROM alpine:latest

WORKDIR /www

COPY --from=builder /build/main /www/
COPY --from=builder /build/.env /www/.env

ENTRYPOINT ["/www/main"]
