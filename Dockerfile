FROM golang:1.14 as builder
LABEL maintainer "Eric Yan <docker@ericyan.me>"

WORKDIR /go/src/github.com/ericyan/omnicast
COPY . .
RUN make build

FROM busybox:glibc
COPY --from=builder /go/src/github.com/ericyan/omnicast/bin/omnicastd /usr/local/bin/
CMD omnicastd
