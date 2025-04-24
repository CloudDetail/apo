FROM golang:1.23.8-bullseye AS builder
LABEL authors="apo"
WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o apo-backend

FROM debian:bullseye-slim AS runner
WORKDIR /app

RUN apt-get update && \
    apt-get install -y sqlite3 ca-certificates && \
    rm -rf /var/lib/apt/lists/*
COPY config/apo.yml /app/config/apo.yml
COPY --from=builder /build/apo-backend /app/
COPY pkg/services/alerts/notification/default.tmpl /app/notification/default.tmpl
COPY sqlscripts /app/sqlscripts
COPY static /app/static

CMD ["/app/apo-backend"]
