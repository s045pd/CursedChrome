FROM golang:1.25-alpine AS build
WORKDIR /src

# Cache deps
COPY go.mod go.sum ./
RUN go mod download

# Build static binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath -ldflags="-s -w" \
    -o /out/cursed-server ./cmd/cursed-server

# Runtime
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata curl wget && \
    adduser -D -H -u 10001 cursed
COPY --from=build /out/cursed-server /usr/local/bin/cursed-server
USER cursed
EXPOSE 8118 4343 8080
HEALTHCHECK --interval=30s --timeout=3s --retries=3 \
    CMD wget -qO- http://127.0.0.1:8118/health | grep -q '"success":true' || exit 1
ENTRYPOINT ["/usr/local/bin/cursed-server"]
