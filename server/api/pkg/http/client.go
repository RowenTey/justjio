package http

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// NewHTTPClient creates an HTTP client with OpenTelemetry instrumentation
func NewHTTPClient() HTTPClient {
	return &http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}
}
