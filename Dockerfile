FROM golang:1.25-alpine3.21 AS toolset

RUN apk add gcc make git musl-dev

ARG GITHUB_TOKEN
ENV GOPRIVATE=github.com/vpnhouse/*

# Configure Git to use token authentication for private modules
RUN git config --global url."https://${GITHUB_TOKEN}@github.com/".insteadOf "https://github.com"


FROM node:20-alpine AS nodejs

RUN corepack enable && corepack prepare pnpm@latest --activate

COPY ./frontend /app/
WORKDIR /app
RUN pnpm install && pnpm run build

FROM toolset AS gomodules
RUN apk add openssh-client
COPY go.mod /build/
COPY .gitconfig /root/
WORKDIR /build
RUN mkdir -p -m 0600 ~/.ssh && ssh-keyscan github.com >> ~/.ssh/known_hosts
RUN --mount=type=ssh go mod download

FROM gomodules AS builder

COPY . /build
COPY --from=nodejs /app/dist /build/internal/frontend/dist/
WORKDIR /build
ARG FEATURE_SET={$FEATURE_SET:-personal}
RUN FEATURE_SET=$FEATURE_SET make build/app


FROM alpine:3.21

RUN apk add tcpdump wireguard-tools nftables
COPY --from=builder /build/tunnel-node /tunnel-node
CMD ["/tunnel-node"]
