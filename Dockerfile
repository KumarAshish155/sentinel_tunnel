FROM golang:alpine AS builder

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /build

# Copy and download dependency using go mod
COPY go.mod .
COPY go.sum .
RUN go mod download

# Copy the code into the container
COPY . .

# Build the application
RUN go build -o tunnel sentinel_tunnelling_client.go

# Move to /dist directory as the place for resulting binary folder
WORKDIR /dist

RUN cp /build/tunnel .

# Build a small image
FROM alpine:3

COPY --from=builder /dist/tunnel /

# Command to run
ENTRYPOINT ["/tunnel"]