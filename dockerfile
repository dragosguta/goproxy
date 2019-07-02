FROM golang:1.12.6-alpine

# Install git
RUN apk update && apk add git

WORKDIR /go/src/app

COPY . .

RUN go get -v ./...
RUN go build -o goproxy

# Use a smaller alpine image
FROM alpine:latest

RUN apk update \
  && apk upgrade \
  && apk add --no-cache \
  ca-certificates \
  && update-ca-certificates 2>/dev/null || true

WORKDIR /home

# Copy the binary file from the first image
COPY --from=0 /go/src/app/goproxy .

CMD ["./goproxy"]