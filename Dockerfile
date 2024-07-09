FROM golang:1.21-alpine3.18 AS toolset

RUN apk add gcc make git musl-dev


FROM node:16-alpine3.14 AS nodejs

COPY ./frontend /app/
WORKDIR /app
RUN npm install && npm run build

FROM toolset AS gomodules
RUN apk add openssh-client
COPY go.mod /build/
COPY .gitconfig /root/
WORKDIR /build
RUN mkdir -p -m 0600 ~/.ssh && ssh-keyscan github.com >> ~/.ssh/known_hosts
RUN --mount=type=ssh GOPRIVATE=github.com/vpnhouse go mod download

FROM gomodules AS builder

COPY . /build
COPY --from=nodejs /app/dist /build/internal/frontend/dist/
WORKDIR /build
ARG FEATURE_SET={$FEATURE_SET:-personal}
RUN FEATURE_SET=$FEATURE_SET make build/app


FROM alpine:3.18

RUN apk add tcpdump wireguard-tools nftables
COPY --from=builder /build/tunnel-node /tunnel-node
CMD ["/tunnel-node"]
