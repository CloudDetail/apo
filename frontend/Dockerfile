FROM 192.168.1.22:81/kindling/node:22.19.0-bookworm AS build

WORKDIR /app/

ENV YARN_CACHE_FOLDER=/usr/local/share/.cache/yarn

COPY package.json yarn.lock ./

# 1. 配置 Yarn 国内源和超时时间
RUN yarn config set registry https://registry.npmmirror.com \
    && yarn config set network-timeout 600000 \
    && yarn config set network-concurrency 8 \
    && yarn config set cache-folder "$YARN_CACHE_FOLDER"

# 2. 验证构建时需要的证书和 git，避免每次构建访问外部 apt 源
RUN git --version \
    && update-ca-certificates

# 3. 替换 yarn.lock 里可能残留的国外 registry
RUN sed -i 's|https://registry.yarnpkg.com|https://registry.npmmirror.com|g' yarn.lock \
    && sed -i 's|https://registry.npmjs.org|https://registry.npmmirror.com|g' yarn.lock

# 4. 安装依赖
RUN --mount=type=cache,id=apo-front-yarn-cache,target=/usr/local/share/.cache/yarn \
    yarn install --frozen-lockfile --network-timeout 600000 --network-concurrency 8 --cache-folder "$YARN_CACHE_FOLDER" \
    || yarn install --frozen-lockfile --network-timeout 600000 --network-concurrency 8 --cache-folder "$YARN_CACHE_FOLDER" --prefer-offline

COPY . .

ARG VITE_PUBLIC_POSTHOG_HOST
ARG VITE_PUBLIC_POSTHOG_KEY
ARG VITE_SHOW_MYSQL_HEALTH

ENV NODE_OPTIONS="--openssl-legacy-provider --max-old-space-size=6144"

RUN yarn build

FROM 192.168.1.22:81/kindling/nginx:stable-alpine3.20

COPY --from=build /app/build /usr/share/nginx/html
COPY ./nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
