package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMain(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping main test with DB dependency in short mode")
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fpath := filepath.Join("testdata", "wikipedia_2020-04-05.htm")
		reader, err := os.Open(fpath)
		require.NoError(t, err)
		_, err = io.Copy(w, reader)
		require.NoError(t, err)
	}))
	defer ts.Close()
	scrapeURL = ts.URL
	buffer := bytes.Buffer{}
	out = &buffer
	main()

	lines := strings.Split(buffer.String(), "\n")
	require.Equal(t, 225, len(lines))
	wantLine0 := `                     country  cases deaths recoveries`
	wantLine1 := `               United States 311616   8489      14943`
	wantLineN := `            Papua New Guinea      1      0          0`
	require.Equal(t, wantLine0, lines[0])
	require.Equal(t, wantLine1, lines[1])
	require.Equal(t, wantLineN, lines[len(lines)-2])
}
