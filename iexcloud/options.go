package iexcloud

import "time"

type Option interface {
	apply(*Client)
}

type optFunc func(*Client)

func (f optFunc) apply(c *Client) {
	f(c)
}

// BatchEndpoint accepts the full URL to the IEX Cloud API to use for batch
// requests.
func BatchEndpoint(url string) Option {
	return optFunc(func(c *Client) {
		c.batchEndpoint = url
	})
}

// CallTimeout accepts a duration that the client should wait for a response
// for each call.
func CallTimeout(timeout time.Duration) Option {
	return optFunc(func(c *Client) {
		c.timeout = timeout
	})
}
