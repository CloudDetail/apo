FROM registry.cn-hangzhou.aliyuncs.com/kindlingx/node:22.19.0-alpine3.21 AS build

WORKDIR /app/

COPY package.json yarn.lock ./

RUN yarn config set registry https://registry.npmmirror.com

# 替换 Alpine apk 源为阿里云镜像，再安装 git
RUN sed -i 's|https://dl-cdn.alpinelinux.org/alpine|https://mirrors.aliyun.com/alpine|g' /etc/apk/repositories \
    && apk add --no-cache ca-certificates git \
    && update-ca-certificates

RUN yarn install

COPY . .

ARG VITE_PUBLIC_POSTHOG_HOST
ARG VITE_PUBLIC_POSTHOG_KEY
ARG VITE_SHOW_MYSQL_HEALTH

# 合并 NODE_OPTIONS，避免前面的 --openssl-legacy-provider 被覆盖
ENV NODE_OPTIONS="--openssl-legacy-provider --max-old-space-size=6144"

RUN yarn build

FROM registry.cn-hangzhou.aliyuncs.com/kindlingx/nginx:alpine

COPY --from=build /app/build /usr/share/nginx/html
COPY ./nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]