package main

import (
	"net/http"

	"github.com/hashicorp/go-retryablehttp"
)

type RoundTripper func(*http.Request) (*http.Response, error)

func (r RoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return r(req)
}

func WrappedRetriableHTTPClient(retriableClient *retryablehttp.Client) *http.Client {
	return &http.Client{
		Transport: RoundTripper(
			func(req *http.Request) (*http.Response, error) {
				retriableRequest, err := retryablehttp.FromRequest(req)

				if err != nil {
					return nil, err
				}

				return retriableClient.Do(retriableRequest)
			},
		),
	}
}
