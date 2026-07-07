FROM node:24-alpine AS build

WORKDIR /app
RUN corepack enable

COPY package.json pnpm-lock.yaml pnpm-workspace.yaml tsconfig.base.json ./
COPY packages ./packages
COPY apps/citizen-web ./apps/citizen-web

RUN pnpm install --frozen-lockfile
RUN pnpm --filter @nadaa/citizen-web build

FROM nginxinc/nginx-unprivileged:1.27-alpine

COPY infra/docker/nginx-spa.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/apps/citizen-web/dist /usr/share/nginx/html

USER 101
EXPOSE 8080
