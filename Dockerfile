# ============================================
# Stage 1: Build GUI
# ============================================
FROM node:18-alpine AS gui-builder
WORKDIR /work/gui
COPY gui/package.json gui/package-lock.json ./
RUN npm ci --omit=dev && npm cache clean --force
COPY gui/ ./
RUN npm run build && rm -rf node_modules

# ============================================
# Stage 2: Production dependencies
# ============================================
FROM node:18-alpine AS deps
WORKDIR /work
COPY package.json package-lock.json ./
RUN npm ci --omit=dev && npm cache clean --force

# ============================================
# Stage 3: Final minimal production image
# ============================================
FROM node:18-alpine AS production

# Install only essential runtime dependencies
RUN apk add --no-cache ffmpeg

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

# Clean up any extra files
RUN rm -rf /root/.npm /tmp/*

ENTRYPOINT ["/work/docker-entrypoint.sh"]
