package utils

import (
	"net/http"

	"github.com/stretchr/testify/mock"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
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
