FROM golang:latest

RUN mkdir /go/src/go-datadog-apm-example
WORKDIR /go/src/go-datadog-apm-example
ADD . /go/src/go-datadog-apm-example
RUN GO111MODULE=on go build .
ENTRYPOINT [ "./go-datadog-apm-example" ]