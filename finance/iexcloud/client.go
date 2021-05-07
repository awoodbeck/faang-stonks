// Package iexcloud provides a finance.Provider implementation that retrieves
// its data from IEX Cloud's API.
package iexcloud

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/awoodbeck/faang-stonks/finance"
)

const (
	// DefaultBatchEndpoint is the default batchQuotes endpoint for the IEX
	// Cloud API.
	DefaultBatchEndpoint = "https://sandbox.iexapis.com/stable/stock/market/batch"

	// DefaultTimeout is the default duration the client waits to a response.
	DefaultTimeout = 10 * time.Second
)

var (
	_ finance.Provider = (*Client)(nil)

	ErrInvalidToken = fmt.Errorf("invalid token")
)

// Client is an IEX Cloud API client.
type Client struct {
	batchEndpoint string
	timeout       time.Duration
	token         string

	httpClient *http.Client
}

// GetQuotes accepts one or more stock symbols and returns the current quote
// for each stock from the IEX Cloud API.
func (c Client) GetQuotes(ctx context.Context, symbols ...string) (
	[]finance.Quote, error) {
	if len(symbols) == 0 {
		return nil, fmt.Errorf("empty symbols")
	}

	v := url.Values{}
	v.Add("types", "quote")
	v.Add("token", c.token)
	v.Add("symbols", strings.ToLower(strings.Join(symbols, ",")))

	callCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(
		callCtx,
		http.MethodGet,
		fmt.Sprintf("%s?%s", c.batchEndpoint, v.Encode()),
		nil,
	)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		buf := new(bytes.Buffer)
		_, err = io.Copy(buf, resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("decoding response: %s", buf)
	}

	b := make(batchQuotes)
	err = json.NewDecoder(resp.Body).Decode(&b)
	if err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return b.MarshalQuotes()
}

// New accepts an API endpoint and returns a pointer to a new Client object.
//
// Defaults:
//     BatchEndpoint = "https://sandbox.iexapis.com/stable/stock/market/batch"
//     CallTimeout   = 10 * time.Second
func New(token string, options ...Option) (*Client, error) {
	if token == "" {
		return nil, ErrInvalidToken
	}

	c := &Client{
		batchEndpoint: DefaultBatchEndpoint,
		httpClient:    http.DefaultClient,
		timeout:       DefaultTimeout,
		token:         token,
	}

	for _, option := range options {
		if option != nil {
			option(c)
		}
	}

	if _, err := url.Parse(c.batchEndpoint); err != nil {
		return nil, fmt.Errorf("batchQuotes endpoint %q: %w", c.batchEndpoint,
			err)
	}

	return c, nil
}
