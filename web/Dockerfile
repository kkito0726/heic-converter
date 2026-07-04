# フロントエンドのイメージ。Viteでビルドした静的ファイルをnginxで配信し、
# RPCパスはapiコンテナへリバースプロキシする(開発時のViteプロキシと同じ構図。
# 同一オリジンになるためCORS設定は不要)。
FROM node:22-alpine AS build
WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci

COPY . .
RUN npm run build

FROM nginx:alpine
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/dist /usr/share/nginx/html
EXPOSE 80
