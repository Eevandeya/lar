package client

import "net/http"

type Client struct {
	baseURL string
	secret  string
	client  *http.Client
}

func New(baseURL string, secret string) *Client {
	return &Client{
		baseURL: baseURL,
		secret:  secret,
		client:  &http.Client{},
	}
}

func (c *Client) url(path string) string {
	return c.baseURL + path
}
