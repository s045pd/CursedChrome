FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata wget && \
    adduser -D -H -u 10001 cursed
COPY cursed-server /usr/local/bin/cursed-server
RUN chmod +x /usr/local/bin/cursed-server
# GUI bundle (Vue dist). Baked in like the original Node image does, so
# the API server can serve / -> index.html with no extra mount.
COPY gui-dist /work/gui/dist
COPY extensions /work/extensions
USER cursed
EXPOSE 8118 4343 8080
HEALTHCHECK --interval=30s --timeout=3s --retries=3 \
    CMD wget -qO- http://127.0.0.1:8118/health | grep -q '"success":true' || exit 1
ENTRYPOINT ["/usr/local/bin/cursed-server"]
