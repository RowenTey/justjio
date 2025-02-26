# Build the application from source
FROM golang:1.23-alpine AS build-stage

# Kafka dependencies
RUN set -ex &&\
    apk add --no-progress --no-cache \
      gcc \
      musl-dev

WORKDIR /app

COPY go.* ./
RUN go mod download

COPY . .
RUN go get -d -v
RUN GOOS=linux GOARCH=amd64 go build -v -tags musl -o /ws-server

# Deploy the application binary into a lean image
# FROM gcr.io/distroless/base-debian11 AS build-release-stage
FROM alpine:latest AS build-release-stage

WORKDIR /

# Download cURL for healthcheck
RUN apk add --no-cache curl

COPY --from=build-stage /ws-server /ws-server

EXPOSE 8081

ENTRYPOINT ["/ws-server"]