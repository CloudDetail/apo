FROM golang:1.21.5-bullseye AS builder
LABEL authors="apo"
WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o apo-backend

FROM debian:bullseye-slim AS runner
WORKDIR /app
ENV TZ=Asia/Shanghai
COPY config/apo.yml /app/config/apo.yml
COPY --from=builder /build/apo-backend /app/
CMD ["/app/apo-backend"]