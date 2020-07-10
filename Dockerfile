FROM golang:1.14.4-alpine3.12 as golang-builder
ENV CGO_ENABLED=0
ENV GO111MODULE=on
WORKDIR /go/src/app
COPY . .
RUN apk --no-cache add make git gcc libtool musl-dev unzip &&\
  make build
# ex. docker build -t secberus/projector:{{ version}}
FROM alpine:3.12 as service
RUN apk --no-cache add ca-certificates && \
    rm -rf /var/cache/apk/* /tmp/*
COPY --from=golang-builder /go/src/app/projector .
ENTRYPOINT ["./projector"]
