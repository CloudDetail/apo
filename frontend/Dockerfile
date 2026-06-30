FROM registry.cn-hangzhou.aliyuncs.com/kindlingx/node:22.19.0-alpine3.21 AS build

WORKDIR /app/

COPY package.json yarn.lock ./

# 1. 配置 Yarn 国内源和超时时间
RUN yarn config set registry https://registry.npmmirror.com \
    && yarn config set network-timeout 600000

# 2. 替换 Alpine apk 源为阿里云镜像，并安装 git
RUN sed -i 's|https://dl-cdn.alpinelinux.org/alpine|https://mirrors.aliyun.com/alpine|g' /etc/apk/repositories \
    && apk add --no-cache ca-certificates git \
    && update-ca-certificates

# 3. 替换 yarn.lock 里可能残留的国外 registry
RUN sed -i 's|https://registry.yarnpkg.com|https://registry.npmmirror.com|g' yarn.lock \
    && sed -i 's|https://registry.npmjs.org|https://registry.npmmirror.com|g' yarn.lock

# 4. 安装依赖
RUN yarn install --frozen-lockfile --network-timeout 600000

COPY . .

ARG VITE_PUBLIC_POSTHOG_HOST
ARG VITE_PUBLIC_POSTHOG_KEY
ARG VITE_SHOW_MYSQL_HEALTH

ENV NODE_OPTIONS="--openssl-legacy-provider --max-old-space-size=6144"

RUN yarn build

FROM registry.cn-hangzhou.aliyuncs.com/kindlingx/nginx:alpine

COPY --from=build /app/build /usr/share/nginx/html
COPY ./nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]