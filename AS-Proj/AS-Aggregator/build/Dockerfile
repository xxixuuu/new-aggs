FROM golang:alpine

RUN apk add --no-cache go \
        binutils-gold \
        curl \
        g++ \
        gcc \
        libgcc \
        linux-headers \
        make \
        openssl-dev
# Set necessary environmet variables needed for our image
ENV GO111MODULE=on \
    CGO_ENABLED=1 \
    GOOS=linux \
    GOARCH=amd64

# Move to working directory /build
WORKDIR /build

# Copy the code into the container
COPY . .

# Copy and download dependency using go mod
COPY go.mod .
COPY go.sum .

RUN go mod download

# This container exposes port 3002 to the outside world
EXPOSE 3000

# Build the application
RUN go build -o amain .

# Move to /dist directory as the place for resulting binary folder
WORKDIR /dist

# Copy binary from build to main folder
RUN cp /build/amain .

RUN cp /build/aconfig.json .

ENTRYPOINT [ "/dist/amain" ]

# Command to run when starting the container
CMD ["echo $@"]