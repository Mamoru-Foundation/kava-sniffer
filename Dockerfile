#FROM golang:1.20-alpine AS build-env
FROM golang:1.20 as build-env

# Set up dependencies
# bash, jq, curl for debugging
# git, make for installation
# libc-dev, gcc, linux-headers, eudev-dev are used for cgo and ledger installation
#RUN apk add bash git make libc-dev gcc linux-headers eudev-dev jq curl
RUN apt-get update  \
    && apt-get install -y gcc make git curl jq git tar bash libc6-dev
# Set working directory for the build
WORKDIR /root/kava
# default home directory is /root

# Copy dependency files first to facilitate dependency caching
COPY ./go.mod ./
COPY ./go.sum ./

# Download dependencies
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    go version && go mod download

# Add source files
COPY . .

#ENV LEDGER_ENABLED False

# Mount go build and mod caches as container caches, persisted between builder invocations
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    make install

#FROM alpine:3.15
FROM debian:12.0-slim

RUN touch /var/run/supervisor.sock

#RUN apk add bash jq curl
RUN apt-get update  \
    && apt-get install -y ca-certificates jq unzip bash grep curl sed htop procps cron supervisor \
    && apt-get clean


COPY --from=build-env /go/bin/kava /bin/kava

CMD ["kava"]
