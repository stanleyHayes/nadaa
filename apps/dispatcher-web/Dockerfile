FROM node:24-alpine AS build

WORKDIR /app
RUN corepack enable

COPY package.json pnpm-lock.yaml pnpm-workspace.yaml tsconfig.base.json ./
COPY packages ./packages
COPY apps/dispatcher-web ./apps/dispatcher-web

RUN pnpm install --frozen-lockfile
RUN pnpm --filter @nadaa/dispatcher-web build

FROM nginx:1.27-alpine

COPY infra/docker/nginx-spa.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/apps/dispatcher-web/dist /usr/share/nginx/html

EXPOSE 8080
