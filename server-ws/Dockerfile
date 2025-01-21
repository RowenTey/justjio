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

# # Run the tests in the container
# FROM build-stage AS run-test-stage
# RUN go test -v ./...

# Deploy the application binary into a lean image
# FROM gcr.io/distroless/base-debian11 AS build-release-stage
FROM alpine:latest AS build-release-stage

WORKDIR /

COPY --from=build-stage /ws-server /ws-server

EXPOSE 8080

ENTRYPOINT ["/ws-server"]