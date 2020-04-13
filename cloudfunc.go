package cloudfunc

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/juliaogris/covid19/pkg/covid19"
)

const conn = "" // connection string parsed from envvars by lib/pq

func Covid19HTTP(w http.ResponseWriter, r *http.Request) {
	t, err := covid19.ScrapeWiki(covid19.WikiURL, conn)
	if err != nil {
		log.Println("Covid19HTTP ERROR:", err)
		fmt.Fprintln(w, "Error", err)
		return
	}
	log.Println("Covid19HTTP: successfully added", len(t.Cells), "rows.")
	fmt.Fprintln(w, "Successfully added", len(t.Cells), "rows.")
}

func Covid19Event(ctx context.Context, _ interface{}) error {
	t, err := covid19.ScrapeWiki(covid19.WikiURL, conn)
	if err != nil {
		log.Println("ConvidEvent ERROR:", err)
		return err
	}
	log.Println("ConvidEvent: successfully added", len(t.Cells), "rows.")
	return nil
}
