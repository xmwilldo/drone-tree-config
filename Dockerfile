FROM golang:1.20-alpine AS builder

RUN true \
  && sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
  && apk add -U --no-cache ca-certificates git binutils

ADD . /go/src/github.com/bitsbeats/drone-tree-config
WORKDIR /go/src/github.com/bitsbeats/drone-tree-config

ENV CGO_ENABLED=0 \
    GO111MODULE=on \
    GOPROXY=https://goproxy.cn,direct

RUN true \
  && go test ./plugin \
  && go build -o drone-tree-config github.com/bitsbeats/drone-tree-config/cmd/drone-tree-config \
  && strip drone-tree-config

# ---

FROM alpine

RUN true \
  && sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
  && apk add -U --no-cache ca-certificates
COPY --from=builder /go/src/github.com/bitsbeats/drone-tree-config/drone-tree-config /usr/local/bin
CMD /usr/local/bin/drone-tree-config
