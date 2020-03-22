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
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fpath := filepath.Join("testdata", "coronavirus-map-2020-03-22.html")
		reader, err := os.Open(fpath)
		require.NoError(t, err)
		_, err = io.Copy(w, reader)
		require.NoError(t, err)
	}))
	defer ts.Close()
	sourceURL = ts.URL
	buffer := bytes.Buffer{}
	out = &buffer
	main()

	lines := strings.Split(buffer.String(), "\n")
	require.Equal(t, 169, len(lines))
	want := `                       Worldwide   303594     43.1    94625    12964`
	require.Equal(t, want, lines[0])
}
