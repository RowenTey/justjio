package utils

import (
	"net/http"

	"github.com/stretchr/testify/mock"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// NewHTTPClient creates an HTTP client with OpenTelemetry instrumentation
func NewHTTPClient() HTTPClient {
	return &http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}
}

type MockHTTPClient struct {
	mock.Mock
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*http.Response), args.Error(1)
}
