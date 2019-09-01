package ravelry

import (
	"errors"
	"net/http"
)

// Endpoint is the top level endpoint for Ravelry's API
const Endpoint = "https://api.ravelry.com/"

// Client is a ravelry.com ravelry that adds Authorization headers
// to all requests
type Client struct {
	http.Client
	username string
	password string
}

func NewClient(username, password string) *Client {
	return &Client{
		Client: http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return errors.New(req.URL.String())
			},
		},
		username: username,
		password: password,
	}
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	req.SetBasicAuth(c.username, c.password)
	return c.Client.Do(req)
}
