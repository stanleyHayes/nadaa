FROM node:24-alpine AS build

WORKDIR /app
RUN corepack enable

COPY package.json pnpm-lock.yaml pnpm-workspace.yaml tsconfig.base.json ./
COPY packages ./packages
COPY apps/authority-dashboard ./apps/authority-dashboard

RUN pnpm install --frozen-lockfile
RUN pnpm --filter @nadaa/authority-dashboard build

FROM nginx:1.27-alpine

COPY infra/docker/nginx-spa.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/apps/authority-dashboard/dist /usr/share/nginx/html

EXPOSE 8080
