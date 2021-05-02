package sqlite

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/awoodbeck/faang-stonks/finance"
)

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
		}
	}
}
