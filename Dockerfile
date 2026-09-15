# syntax=docker/dockerfile:1.7
# Hanzo Ads — the Vite SPA built once, served by hanzoai/spa.
FROM node:24-alpine AS build
WORKDIR /build
ENV CI=true PNPM_HOME=/pnpm PATH=$PNPM_HOME:$PATH
RUN corepack enable && corepack prepare pnpm@11.24.0 --activate

# pnpm-workspace.yaml holds the catalog and the one-gui-runtime overrides.
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
RUN pnpm install --frozen-lockfile

COPY tsconfig.json vite.config.ts index.html ./
COPY src src

# `pnpm build` typechecks first. An empty vite build still exits 0, hence the tests.
RUN pnpm build \
 && [ -s dist/index.html ] \
 && [ -d dist/assets ]

# hanzoai/spa answers index.html for every route.
FROM ghcr.io/hanzoai/spa:1.4.11
COPY --from=build /build/dist /public
EXPOSE 3000
