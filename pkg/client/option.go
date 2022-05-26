package client

type Option func(c *Client)

func WithMetrics() Option {
	return func(c *Client) {
		c.metricsEnabled = true
	}
}

func WithRetry() Option {
	return func(c *Client) {
		c.retryEnabled = true
	}
}
