# syntax=docker/dockerfile:1.4
FROM --platform=$BUILDPLATFORM golang:alpine AS builder
WORKDIR /src
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 go build -o /out/wwpm main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /out/wwpm .
COPY config.toml .
EXPOSE 8080
ENTRYPOINT ["./wwpm"]