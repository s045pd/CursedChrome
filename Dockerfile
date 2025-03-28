FROM node:12-slim AS gui-builder
COPY gui/package.json gui/package-lock.json /work/gui/
RUN cd /work/gui && npm ci && npm cache clean --force
COPY gui /work/gui
RUN cd /work/gui && npm run build && npm cache clean --force


FROM node:12-slim AS production
RUN sed -i 's/deb.debian.org/archive.debian.org/g' /etc/apt/sources.list && \
	sed -i 's/security.debian.org/archive.debian.org/g' /etc/apt/sources.list && \
	sed -i '/stretch-updates/d' /etc/apt/sources.list && \
	apt-get update && \
	apt-get install -y --no-install-recommends ffmpeg && \
	apt-get clean && \
	rm -rf /var/lib/apt/lists/*

WORKDIR /work/
COPY package.json package-lock.json .
RUN npm ci && npm cache clean --force

COPY anyproxy/ ./anyproxy/
COPY --from=gui-builder /work/gui/dist /work/gui/dist

COPY utils.js api-server.js server.js database.js docker-entrypoint.sh .
ENTRYPOINT ["/work/docker-entrypoint.sh"]
