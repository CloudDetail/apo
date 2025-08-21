FROM golang:1.23.8-bookworm AS builder
LABEL authors="apo"
WORKDIR /build

ENV GOPROXY https://goproxy.cn
COPY go.mod go.sum ./
RUN go mod download && go mod verify

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
