ARG ARCH=amd64

FROM golang:1.14 as builder
LABEL maintainer "Eric Yan <docker@ericyan.me>"

ARG ARCH
WORKDIR /go/src/github.com/ericyan/omnicast
COPY . .
RUN make bin/omnicastd-${ARCH}

FROM scratch
ARG ARCH
COPY --from=builder /go/src/github.com/ericyan/omnicast/bin/omnicastd-${ARCH} /usr/local/bin/omnicastd
ENTRYPOINT ["/usr/local/bin/omnicastd"]
