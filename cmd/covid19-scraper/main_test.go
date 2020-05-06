package main

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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
	*conn = "postgres://postgres:postgres@localhost:5432/?sslmode=disable"

	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	main()

	os.Stdout = stdout
	w.Close()
	out, _ := ioutil.ReadAll(r)

	require.Equal(t, "Successfully added 220 rows.\n", string(out))
}
