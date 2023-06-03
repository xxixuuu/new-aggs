FROM golang:alpine

# Set necessary environmet variables needed for our image
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

RUN apk add --update --no-cache vim git make musl-dev go curl
RUN export GOPATH=/root/go
RUN export PATH=${GOPATH}/bin:/usr/local/go/bin:$PATH
RUN export GOBIN=$GOROOT/bin
RUN mkdir -p ${GOPATH}/src ${GOPATH}/bin
RUN export GO111MODULE=on
RUN go version

WORKDIR /build
RUN git clone https://github.com/udhos/goben ;
WORKDIR /build/goben
RUN go test ./goben
RUN CGO_ENABLED=0 go install ./goben

EXPOSE 8080

ENTRYPOINT [ "/go/bin/goben" ]
