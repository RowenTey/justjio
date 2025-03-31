# Build the application from source
FROM golang:1.23-alpine AS build-stage

ARG GOARCH=amd64

# Kafka dependencies
RUN set -ex &&\
  apk add --no-progress --no-cache \
  gcc \
  musl-dev

WORKDIR /app

COPY go.* ./
RUN go mod download

COPY . .
RUN GOOS=linux GOARCH=${GOARCH} go build -v -tags musl -o /ws-server

# Deploy the application binary into a lean image
FROM alpine:latest AS build-release-stage

WORKDIR /

# Download cURL for healthcheck
RUN apk add --no-cache curl

COPY --from=build-stage /ws-server /ws-server

EXPOSE 8081

ENTRYPOINT ["/ws-server"]