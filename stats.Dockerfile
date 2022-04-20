FROM golang:1.17.6-alpine3.15 as toolset
RUN apk add gcc musl-dev

FROM toolset as builder
WORKDIR /code
COPY . /code
RUN go mod tidy
RUN go build -o app cmd/statserver/main.go


FROM alpine:3.15
COPY --from=builder /code/app /app
RUN mkdir /extstat-data/
CMD ["/app"]
