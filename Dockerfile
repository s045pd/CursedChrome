# ============================================
# Stage 1: Build Vue 3 frontend
# ============================================
FROM node:22-alpine AS gui-builder
WORKDIR /src/gui-next
COPY gui-next/package.json gui-next/package-lock.json ./
RUN npm ci && npm cache clean --force
COPY gui-next/ ./
RUN npm run build

# ============================================
# Stage 2: Build Go backend
# ============================================
FROM golang:1.25-alpine AS go-builder
WORKDIR /src
COPY cursed-go/go.mod cursed-go/go.sum ./
RUN go mod download
COPY cursed-go/ ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -trimpath -ldflags="-s -w" \
    -o /out/cursed-server ./cmd/cursed-server

# ============================================
# Stage 3: Minimal production image
# ============================================
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata wget \
    && adduser -D -H -u 10001 cursed

COPY --from=go-builder /out/cursed-server /usr/local/bin/cursed-server
COPY --from=gui-builder /src/gui/dist /work/gui/dist

ENV GUI_DIST_PATH=/work/gui/dist

USER cursed
EXPOSE 8118 4343 8080

HEALTHCHECK --interval=30s --timeout=3s --retries=3 \
    CMD wget -qO- http://127.0.0.1:8118/health | grep -q '"success":true' || exit 1

ENTRYPOINT ["/usr/local/bin/cursed-server"]
