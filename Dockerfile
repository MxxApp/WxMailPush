# syntax=docker/dockerfile:1.4
FROM --platform=$BUILDPLATFORM golang:alpine AS builder
WORKDIR /src
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 go build -o /out/wxmailpush main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /out/wxmailpush .
COPY config.toml .
EXPOSE 8080
ENTRYPOINT ["./wxmailpush", "-config=config.toml"]