package cloudsearch_test

import (
	"github.com/herval/cloudsearch"
	"reflect"
	"testing"
	"time"
)

func TestParser(t *testing.T) {
	q := "type:File foo bar service:dropbox before:2017-01-1 service:invalid service:Google x"

	parsed := cloudsearch.ParseQuery(q, "1")
	if parsed.RawText != q {
		t.Fatal()
	}

	if *parsed.Before != time.Date(2017, 1, 1, 0, 0, 0, 0, time.UTC) {
		t.Fatal(parsed)
	}

	if parsed.After != nil {
		t.Fatal(parsed)
	}

	if !reflect.DeepEqual(parsed.AccountTypes, []AccountType{Dropbox, Google}) {
		t.Fatal(parsed)
	}

	if !reflect.DeepEqual(parsed.ContentTypes, []ContentType{File}) {
		t.Fatal(parsed)
	}

	if parsed.Text != "foo bar     x" {
		t.Fatal(parsed.Text)
	}

}
