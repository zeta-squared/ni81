package http

import "net/http"

type authTransport struct {
	apiKey string
	base   http.RoundTripper
}

func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.apiKey == "" {
		return t.base.RoundTrip(req)
	}

	req = req.Clone(req.Context())
	req.Header.Set("Authorization", "Bearer "+t.apiKey)
	return t.base.RoundTrip(req)
}

func NewAuthClient(apiKey string) http.Client {
	return http.Client{
		Transport: &authTransport{
			apiKey: apiKey,
			base:   http.DefaultTransport,
		},
	}
}
