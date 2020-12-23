package client

import "net/http"

// HTTPHandler is just an alias to http.RoundTriper.RoundTrip function
type HTTPHandler func(*http.Request) (*http.Response, error)

// HTTPClientInterceptor is a user defined function that can alter a request before it's sent
// and/or alter a response before it's returned to the caller
type HTTPClientInterceptor func(*http.Request, HTTPHandler) (*http.Response, error)

// HTTPClientBuilder is a builder interface to build http.Client with interceptors
type HTTPClientBuilder interface {
	AddInterceptors(...HTTPClientInterceptor) HTTPClientBuilder
	WithPreconfiguredClient(*http.Client) HTTPClientBuilder
	Build() HttpClient
}

// NewHTTPClientBuilder REST HTTP builder
//
// Useful when you want to create several *http.Client with different options
type NewHTTPClientBuilder func() HTTPClientBuilder
