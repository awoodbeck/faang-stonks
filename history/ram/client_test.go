package ram

import (
	"reflect"
	"strings"
	"testing"

	"github.com/awoodbeck/faang-stonks/finance"
)

func TestNewClient(t *testing.T) {
	t.Parallel()

	c, err := New()
	if err != nil {
		t.Fatal(err)
	}

	if len(defaultSymbols) != len(c.quotes) {
		t.Errorf("the quotes map length mismatches the default symbols slice length")
	}

	for _, symbol := range defaultSymbols {
		if _, ok := c.quotes[strings.ToLower(symbol)]; !ok {
			t.Errorf("%q not found in the quotes map", symbol)
		}
	}
}

func TestNewClientOptions(t *testing.T) {
	t.Parallel()

	symbols := []string{"foo", "bar"}
	c, err := New(Symbols(symbols))
	if err != nil {
		t.Fatal(err)
	}

	if len(symbols) != len(c.quotes) {
		t.Errorf("the quotes map length mismatches the optional symbols slice length")
	}

	for _, symbol := range symbols {
		if _, ok := c.quotes[strings.ToLower(symbol)]; !ok {
			t.Errorf("%q not found in the quotes map", symbol)
		}
	}
}

func TestArchiverProvider(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		quotes   []finance.Quote
		symbol   string
		last     int
		expected []finance.Quote
	}{
		{ // add multiple quotes but return the last one added
			quotes: []finance.Quote{
				{Price: 123.45, Symbol: "fb"},
				{Price: 123.42, Symbol: "fb"},
			},
			symbol: "fb",
			last:   0,
			expected: []finance.Quote{
				{Price: 123.42, Symbol: "fb"},
			},
		},
	}

	c, err := New()
	if err != nil {
		t.Fatal(err)
	}

	for i, tc := range testCases {
		err = c.SetQuotes(nil, tc.quotes)
		if err != nil {
			t.Errorf("%d: set actual: %v", i, err)
			continue
		}

		actual, err := c.GetQuotes(nil, tc.symbol, tc.last)
		if err != nil {
			t.Errorf("%d: get quotes: %v", i, err)
			continue
		}

		if !reflect.DeepEqual(actual, tc.expected) {
			t.Errorf("%d: actual quotes not equal to expected", i)
			t.Logf("expected: %#v", tc.expected)
			t.Logf("actual:   %#v", actual)
			continue
		}
	}
}
