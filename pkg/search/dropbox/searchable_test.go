package dropbox_test

import (
	"context"
	"fmt"
	"github.com/herval/cloudsearch/pkg/search/dropbox"
	"os"
	"testing"
	"time"

	"github.com/herval/cloudsearch/pkg"
)

func TestDropbox(t *testing.T) {
	if os.Getenv("DROPBOX_TOKEN") == "" {
		t.Log("Skipping d test (no token set)")
		t.Skipped()
	}

	d := dropbox.NewSearch(
		cloudsearch.AccountData{
			ID:    "123",
			Token: os.Getenv("DROPBOX_TOKEN"),
		},
	)

	var fetched *cloudsearch.Result = nil

	go func() {
		time.Sleep(time.Second * 10)
		if fetched == nil {
			t.Fatal("No data found")
		}
	}()

	data := d(cloudsearch.ParseQuery("clear_a.gif", "id", cloudsearch.NewRegistry()), context.Background())

	fmt.Println(<-data)
}
