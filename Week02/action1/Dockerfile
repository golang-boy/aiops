FROM golang:1.20.3 as builder
WORKDIR /app
RUN wget https://gist.github.com/abhishekkr/3beebbc1db54b3b54914/raw/16f13fe1d4d222df0ff21db0e1265c5b583c7477/tcp_server.go && CGO_ENABLED=0 go build -o tcp_server tcp_server.go

FROM alpine:latest

ENV TZ Asia/Shanghai
RUN apk add tzdata && cp /usr/share/zoneinfo/${TZ} /etc/localtime && echo ${TZ} > /etc/timezone && apk del tzdata

WORKDIR /
COPY --from=builder /app/tcp_server tcp_server
CMD ["/tcp_server"]
