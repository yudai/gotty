FROM golang:1.8
MAINTAINER Rafael Bodill <justrafi@gmail.com>

ENV TERM=xterm-256color
ENV GOTTY_PORT=8080
ENV GOTTY_ADDRESS=127.0.0.1
EXPOSE $GOTTY_PORT

RUN [ -d /go/src/github.com/yudai/gotty ] \
  || mkdir -p /go/src/github.com/yudai/gotty

WORKDIR /go/src/github.com/yudai/gotty

ENTRYPOINT ["/go/bin/gotty", "--permit-write"]
CMD ["bash"]

COPY . /go/src/github.com/yudai/gotty
RUN go install -v github.com/yudai/gotty
