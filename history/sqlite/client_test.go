package sqlite

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/awoodbeck/faang-stonks/finance"
)

func TestGetQuotes(t *testing.T) {
	t.Parallel()

	now := time.Now()
	testCases := []struct {
		quotes   []finance.Quote
		symbol   string
		last     int
		expected []finance.Quote
	}{
		{ // add multiple quotes but return the last one added
			quotes: []finance.Quote{
				{Price: 123.45, Symbol: "fb", Time: now},
				{Price: 123.42, Symbol: "fb", Time: now},
			},
			symbol: "fb",
			last:   0,
			expected: []finance.Quote{
				{Price: 123.42, Symbol: "fb", Time: now},
			},
		},
	}

	dir, err := ioutil.TempDir("", "stonks")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Logf("removing temp dir: %v", err)
		}
	}()

	t.Logf("using temp directory %q", dir)

	c, err := New(DatabaseFile(filepath.Join(dir, "stonks.sqlite")))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = c.Close() }()

	for i, tc := range testCases {
		err = c.SetQuotes(context.Background(), tc.quotes)
		if err != nil {
			t.Errorf("%d: set actual: %v", i, err)
			continue
		}

		actual, err := c.GetQuotes(context.Background(), tc.symbol, tc.last)
		if err != nil {
			t.Errorf("%d: get quotes: %v", i, err)
			continue
		}

		if len(actual) != len(tc.expected) {
			t.Errorf("%d: actual quote count not equal to expected count", i)
			continue
		}

		for j, q := range actual {
			if expected := tc.expected[j]; q.Price != expected.Price {
				t.Errorf("%d.%d: actual price: %.2f; expected: %.2f", i, j,
					q.Price, expected.Price)
			}
			if expected := tc.expected[j]; q.Symbol != expected.Symbol {
				t.Errorf("%d.%d: actual symbol: %q; expected: %q", i, j,
					q.Symbol, expected.Symbol)
			}
		}
	}
}

func TestGetQuotesBatch(t *testing.T) {
	t.Parallel()

	now := time.Now()
	testCases := []struct {
		quotes   []finance.Quote
		symbols  []string
		last     int
		expected finance.QuoteBatch
	}{
		{ // add multiple quotes but return the last one added
			quotes: []finance.Quote{
				{Price: 123.45, Symbol: "fb", Time: now},
				{Price: 123.42, Symbol: "fb", Time: now},
				{Price: 123.40, Symbol: "fb", Time: now},
				{Price: 234.56, Symbol: "goog", Time: now},
				{Price: 234.51, Symbol: "goog", Time: now},
			},
			symbols: []string{"fb", "goog"},
			last:    2,
			expected: finance.QuoteBatch{
				"fb": {
					{Price: 123.42, Symbol: "fb"},
					{Price: 123.40, Symbol: "fb"},
				},
				"goog": {
					{Price: 234.56, Symbol: "goog"},
					{Price: 234.51, Symbol: "goog"},
				},
			},
		},
	}

	dir, err := ioutil.TempDir("", "stonks")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Logf("removing temp dir: %v", err)
		}
	}()

	t.Logf("using temp directory %q", dir)

	c, err := New(DatabaseFile(filepath.Join(dir, "stonks.sqlite")))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = c.Close() }()

	for i, tc := range testCases {
		err = c.SetQuotes(context.Background(), tc.quotes)
		if err != nil {
			t.Errorf("%d: set actual: %v", i, err)
			continue
		}

		actual, err := c.GetQuotesBatch(context.Background(), tc.symbols,
			tc.last)
		if err != nil {
			t.Errorf("%d: get quotes: %v", i, err)
			continue
		}

		t.Logf("%#v", actual)

		if len(actual) != len(tc.expected) {
			t.Errorf("%d: actual quote count not equal to expected count", i)
			continue
		}
	}
}
