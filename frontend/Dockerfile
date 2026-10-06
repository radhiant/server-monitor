# syntax=docker/dockerfile:1

# ---------------------------------------------------------------------------
# Stage 1 — build the static bundle
#
# Node is only needed here. The published image carries nothing but Nginx and
# the compiled assets, which keeps the central server free of a Node runtime,
# exactly as the zero-dependency deployment model intends.
# ---------------------------------------------------------------------------
FROM node:22-alpine AS build

WORKDIR /app

# Copy manifests first so `npm ci` is cached until dependencies actually change.
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci

COPY frontend/ ./

# `npm run build` runs vue-tsc before vite, so a type error fails the image
# build rather than shipping a broken dashboard.
RUN npm run build


# ---------------------------------------------------------------------------
# Stage 2 — serve
# ---------------------------------------------------------------------------
FROM nginx:1.27-alpine

# The official image renders every *.template through envsubst at startup and
# writes the result into conf.d, which is how the agent addresses in .env reach
# the proxy without rebuilding the image.
COPY deploy/docker/default.conf.template /etc/nginx/templates/default.conf.template

COPY --from=build /app/dist /usr/share/nginx/html

EXPOSE 80

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget -q --spider http://127.0.0.1/ || exit 1
