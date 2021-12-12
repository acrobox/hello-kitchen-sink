FROM golang:alpine AS builder
ENV CGO_ENABLED=0
COPY . /build/
WORKDIR /build
RUN go build -a -installsuffix docker -ldflags='-w -s' -o /build/bin/hello-kitchen-sink /build

FROM ghcr.io/acrobox/docker/minimal:latest
EXPOSE 8080
COPY --from=builder /build/bin/hello-kitchen-sink /usr/local/bin/hello-kitchen-sink
USER user
ENTRYPOINT ["/usr/local/bin/hello-kitchen-sink"]

LABEL org.opencontainers.image.source https://github.com/acrobox/hello-kitchen-sink
