FROM golang:alpine

# Set necessary environmet variables needed for our image
ENV GO111MODULE=on \
    CGO_ENABLED=1 \
    GOOS=linux \
    GOARCH=amd64

RUN apk update

RUN apk --update add alpine-sdk

RUN apk --update add openssl-dev

# Move to working directory /build
WORKDIR /build

# Copy the code into the container
COPY . .

# Copy and download dependency using go mod
COPY go.mod .

#COPY go.sum .
RUN go mod download

# Build the application
RUN go build -o dmain .

# Move to /dist directory as the place for resulting binary folder
WORKDIR /dist

# Copy binary from build to main folder
RUN cp /build/dmain .

RUN cp /build/dconfig.json .

ENTRYPOINT [ "/dist/dmain" ]

# Command to run when starting the container
CMD ["echo $@"]