FROM golang:1.20 as builder

WORKDIR /usr/src/app

RUN go env -w GO111MODULE=auto \
  && go env -w CGO_ENABLED=0 \
  && go env -w GOPROXY=https://goproxy.cn,direct

COPY . .

RUN go mod tidy

RUN set -ex \
    && cd /usr/src/app \
    && go build -ldflags "-s -w -extldflags '-static'" -o ticket

FROM alpine:latest

ENV TZ Asia/Shanghai
RUN apk add alpine-conf && \
    /sbin/setup-timezone -z Asia/Shanghai && \
    apk del alpine-conf

COPY --from=builder /usr/src/app/ticket /usr/bin/ticket
COPY --from=builder /usr/src/app/log.txt /data/log.txt
RUN chmod +x /usr/bin/ticket

WORKDIR /data
