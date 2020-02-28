FROM golang:1.13.1

WORKDIR /gotty
COPY . /gotty
RUN CGO_ENABLED=0 make

FROM alpine:latest

RUN apk update && \
    apk upgrade && \
    apk --no-cache add ca-certificates && \
    apk add bash
WORKDIR /root
COPY --from=0 /gotty/gotty /usr/bin/
CMD ["gotty",  "-w", "bash"]
