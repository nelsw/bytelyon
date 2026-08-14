FROM golang:1.25-alpine AS builder

LABEL maintainer="Connor Van Elswyk"

ENV GO111MODULE=on \
    CGO_ENABLED=0

WORKDIR /app

# Copy and download dependencies (leverages Docker caching)
COPY apps/mux/go.mod apps/mux/go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY apps/mux/ .

# Build a statically linked binary for Linux
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o myapp ./cmd/app/main.go

# --- Stage 2: Final runtime image ---
FROM alpine:3.20

# Add a non-root user for security
RUN adduser -D appuser
USER appuser

WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=builder /app/myapp .

# Expose your application port (change 8080 if needed)
EXPOSE 3000

# Run the binary
ENTRYPOINT ["./myapp"]
