FROM golang:1.14.2-alpine as micro-builder
ENV CGO_ENABLED=0
ENV GO111MODULE=on
WORKDIR /go/src/app
COPY . .
RUN apk --no-cache add make git gcc libtool musl-dev unzip &&\
  go get github.com/micro/protoc-gen-micro/v2 &&\
  go get -u github.com/golang/protobuf/protoc-gen-go &&\
  go get -u github.com/golang/protobuf/proto &&\
  apk add protobuf &&\
  make build
# ex. docker build -t gcr.io/secberus-infrastructure/metrics-service:{{ version}}
FROM alpine:latest as service
RUN apk --no-cache add ca-certificates && \
    rm -rf /var/cache/apk/* /tmp/*
COPY --from=micro-builder /go/src/app/metrics-service .
ENTRYPOINT ["./metrics-service"]
