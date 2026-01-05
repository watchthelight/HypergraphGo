# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /src
COPY go.mod ./
COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /hg ./cmd/hg

# Final stage
FROM alpine:3.20

RUN apk --no-cache add ca-certificates

COPY --from=builder /hg /usr/local/bin/hg

ENTRYPOINT ["hg"]
