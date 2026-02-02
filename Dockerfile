# ============================================
# Stage 1: Build GUI
# ============================================
FROM node:12-alpine AS gui-builder
WORKDIR /work/gui
COPY gui/package.json gui/package-lock.json ./
RUN npm install --production=false && npm cache clean --force
COPY gui/ ./
RUN npm run build && rm -rf node_modules

# ============================================
# Stage 2: Production dependencies
# ============================================
FROM node:12-alpine AS deps
WORKDIR /work
COPY package.json package-lock.json ./
# Install build dependencies for native modules (bcrypt)
RUN apk add --no-cache python2 make g++
# Use npm install instead of npm ci (lock file may be out of sync)
RUN npm install --production --legacy-peer-deps && npm cache clean --force

# ============================================
# Stage 3: Final minimal production image
# ============================================
FROM node:12-alpine AS production

# Install only essential runtime dependencies
# python and build tools needed for some native modules
RUN apk add --no-cache ffmpeg python2 make g++

WORKDIR /work

# Copy production node_modules from deps stage
COPY --from=deps /work/node_modules ./node_modules

# Copy built GUI
COPY --from=gui-builder /work/gui/dist ./gui/dist

# Copy anyproxy library
COPY anyproxy/ ./anyproxy/

# Copy only essential application files
COPY utils.js api-server.js server.js database.js docker-entrypoint.sh package.json ./

# Ensure entrypoint is executable
RUN chmod +x /work/docker-entrypoint.sh

# Clean up build dependencies to reduce size
RUN apk del python2 make g++ && \
    rm -rf /root/.npm /tmp/* /var/cache/apk/*

ENTRYPOINT ["/work/docker-entrypoint.sh"]
