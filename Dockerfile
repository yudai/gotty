FROM golang:1.5

ADD . /go/src/github.com/yudai/gotty
WORKDIR /go/src/github.com/yudai/gotty

# Install tools
RUN go get github.com/jteeuwen/go-bindata/...
RUN go get github.com/tools/godep

# Checkout hterm
RUN git submodule sync && git submodule update --init --recursive

# Restore libraries in Godeps
RUN godep restore

# Build
RUN make
