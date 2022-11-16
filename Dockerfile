FROM golang:1.17.6-alpine3.15 as toolset

RUN apk add gcc make git musl-dev


FROM node:16-alpine3.14 as nodejs

COPY ./frontend /app/
WORKDIR /app
RUN npm install && npm run build

from toolset as gomodules
COPY go.mod /build/
WORKDIR /build
RUN go mod download

FROM gomodules as builder

COPY . /build
COPY --from=nodejs /app/dist /build/internal/frontend/dist/
WORKDIR /build
ARG FEATURE_SET={$FEATURE_SET:-personal}
RUN FEATURE_SET=$FEATURE_SET make build/app


FROM alpine:3.15

RUN apk add tcpdump wireguard-tools nftables
COPY --from=builder /build/tunnel-node /tunnel-node
CMD ["/tunnel-node"]
