# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /src
COPY go.mod ./
COPY . .

ARG VERSION=dev
ARG COMMIT=unknown
ARG DATE=unknown

RUN CGO_ENABLED=0 go build \
      -ldflags="-s -w -X github.com/watchthelight/HypergraphGo/internal/version.Version=${VERSION} -X github.com/watchthelight/HypergraphGo/internal/version.Commit=${COMMIT} -X github.com/watchthelight/HypergraphGo/internal/version.Date=${DATE}" \
      -o /hg ./cmd/hg && \
    CGO_ENABLED=0 go build \
      -ldflags="-s -w -X github.com/watchthelight/HypergraphGo/internal/version.Version=${VERSION} -X github.com/watchthelight/HypergraphGo/internal/version.Commit=${COMMIT} -X github.com/watchthelight/HypergraphGo/internal/version.Date=${DATE}" \
      -o /hottgo ./cmd/hottgo

# Final stage
FROM alpine:3.20

RUN apk --no-cache add ca-certificates

COPY --from=builder /hg /usr/local/bin/hg
COPY --from=builder /hottgo /usr/local/bin/hottgo

ENTRYPOINT ["hg"]
