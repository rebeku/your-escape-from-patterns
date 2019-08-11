package client

import "net/http"

// Client is a ravelry.com client that adds Authorization headers
// to all requests
type Client struct {
	http.Client
	username string
	password string
}

func NewClient(username, password string) *Client {
	return &Client{
		Client:   http.Client{},
		username: username,
		password: password,
	}
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	req.SetBasicAuth(c.username, c.password)
	return c.Client.Do(req)
}
