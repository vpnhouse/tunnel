FROM golang:1.16-alpine3.12 as toolset

RUN apk add gcc make git musl-dev

FROM toolset as builder

ARG TARGET=build/linux

COPY . /build
WORKDIR /build
RUN make $TARGET

FROM alpine:3.12

RUN apk add tcpdump wireguard-tools
COPY docker/tunnel/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

COPY --from=builder /build/tunnel-node /usr/local/bin/tunnel-node

CMD ["/entrypoint.sh"]
