FROM golang:1.17.6-alpine3.15 as toolset

RUN apk add gcc make git musl-dev

FROM toolset as builder

ARG TARGET=build/linux

COPY . /build
WORKDIR /build
RUN make $TARGET

FROM alpine:3.15

RUN apk add tcpdump wireguard-tools
COPY docker/tunnel/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

COPY --from=builder /build/tunnel-node /usr/local/bin/tunnel-node

CMD ["/entrypoint.sh"]
