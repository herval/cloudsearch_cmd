package dropbox

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/herval/cloudsearch"
)

func TestDropbox(t *testing.T) {
	if os.Getenv("DROPBOX_TOKEN") == "" {
		t.Log("Skipping dropbox test (no token set)")
		t.Skipped()
	}

	dropbox := NewSearch(
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

	data := dropbox(cloudsearch.ParseQuery("clear_a.gif", "id"), context.Background())

	fmt.Println(<-data)
}
