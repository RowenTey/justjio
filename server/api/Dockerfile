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
RUN GOOS=linux GOARCH=${GOARCH} go build -v -tags musl -o /api-server

# Deploy the application binary into a lean image
FROM alpine:latest AS build-release-stage

WORKDIR /

# Download cURL for healthcheck
RUN apk add --no-cache curl

COPY --from=build-stage /api-server /api-server

EXPOSE 8080

ENTRYPOINT ["/api-server"]