FROM golang:1.12.5 as build

ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.io
ADD . /go/src/github.com/huangzhiran/host-feature-discovery

WORKDIR /go/src/github.com/huangzhiran/host-feature-discovery

ARG HFD_VERSION=0.0.0
RUN go install -ldflags "-X main.Version=${HFD_VERSION}" github.com/huangzhiran/host-feature-discovery

RUN go test .

FROM huangzhiran/ubuntu-aliyun:18.04

COPY --from=build /go/bin/host-feature-discovery /usr/bin/host-feature-discovery

ENTRYPOINT ["/usr/bin/host-feature-discovery"]
