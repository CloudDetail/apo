FROM golang:1.23.8-bookworm AS builder
LABEL authors="apo"
WORKDIR /build

ARG PMODULE_HOST=''
ARG ACCESS=''

ENV GOPROXY=https://proxy.golang.org|https://mirrors.aliyun.com/goproxy/|https://goproxy.cn|direct
ENV GOPRIVATE $PMODULE_HOST
ENV GONOPROXY $PMODULE_HOST
ENV GOINSECURE $PMODULE_HOST

RUN if [ "$ACCESS" != "" ]; then git config --global url."http://$ACCESS@$PMODULE_HOST".insteadOf "http://$PMODULE_HOST" ; fi


COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download && go mod verify

COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build go build -v -o apo-backend

FROM debian:bookworm-slim AS runner
WORKDIR /app
RUN sed -i 's/deb.debian.org/mirrors.ustc.edu.cn/g' /etc/apt/sources.list.d/debian.sources && \
    sed -i 's/security.debian.org/mirrors.ustc.edu.cn/g' /etc/apt/sources.list.d/debian.sources

RUN apt-get update && \
    apt-get install -y sqlite3 ca-certificates && \
    rm -rf /var/lib/apt/lists/*
COPY config/apo.yml /app/config/apo.yml
COPY --from=builder /build/apo-backend /app/
COPY pkg/receiver/dingtalk/notification/default.tmpl /app/notification/default.tmpl
COPY sqlscripts /app/sqlscripts
COPY static /app/static

CMD ["/app/apo-backend"]
