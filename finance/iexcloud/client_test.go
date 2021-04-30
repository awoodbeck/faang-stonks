package iexcloud

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func TestNewClientDefaults(t *testing.T) {
	t.Parallel()

	_, err := New("")
	if err != ErrInvalidToken {
		t.Errorf("expected: ErrInvalidToken; actual: %v", err)
	}

	c, err := New("stonks!")
	if err != nil {
		t.Error(err)
	}
	if c != nil {
		if c.batchEndpoint != defaultBatchEndpoint {
			t.Errorf("expected endpoint: %q; actual endpoint: %q",
				defaultBatchEndpoint, c.batchEndpoint)
		}
		if c.timeout != defaultTimeout {
			t.Errorf("expected timeout: %v; actual timeout: %v",
				defaultTimeout, c.timeout)
		}
		if c.httpClient == nil {
			t.Error("nil HTTP client")
		}
	}
}

func TestNewClientOptions(t *testing.T) {
	t.Parallel()

	endpoint := "https://nonexistent.domain"
	timeout := 42 * time.Second

	c, err := New(
		"to the moon!",
		BatchEndpoint(endpoint),
		CallTimeout(timeout),
		InstrumentHTTPClient(),
	)
	if err != nil {
		t.Error(err)
	}
	if c != nil {
		if c.batchEndpoint != endpoint {
			t.Errorf("expected endpoint: %q; actual endpoint: %q",
				endpoint, c.batchEndpoint)
		}
		if c.timeout != timeout {
			t.Errorf("expected timeout: %v; actual timeout: %v",
				timeout, c.timeout)
		}
		if c.httpClient == nil {
			t.Error("nil HTTP client")
		} else {
			if _, ok := c.httpClient.Transport.(promhttp.RoundTripperFunc); !ok {
				t.Error("underlying HTTP transport is not instrumented")
			}
		}
	}
}

func TestNewClientInvalidEndpoint(t *testing.T) {
	t.Parallel()

	_, err := New("this is fine", BatchEndpoint("blah\n"))
	if err == nil {
		t.Errorf("expected a batchQuotes endpoint error; actual: %q", err)
	}
}

// TODO: Mock up an end point that serves up IEX Cloud quotes and test the
// client.
