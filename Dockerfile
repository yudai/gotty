FROM golang:1.13 as builder
ENV GO111MODULE=on
ENV GOPROXY https://goproxy.io
ENV CGO_ENABLED 0
ENV GOOS linux
ENV GO15VENDOREXPERIMENT 1
ENV GOFLAGS -mod=vendor

WORKDIR /app

#RUN go mod download
COPY . .
RUN go build -o /bin/gotty *.go

FROM golang:1.13 as runner
WORKDIR /app

COPY --from=builder /bin/gotty  /bin/gotty
EXPOSE 8000

VOLUME /data

CMD ["gotty"]
