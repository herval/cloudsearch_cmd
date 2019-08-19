package cloudsearch_test

import (
	"github.com/herval/cloudsearch/pkg"
	"github.com/herval/cloudsearch/pkg/test"
	"reflect"
	"testing"
	"time"
)

func TestParser(t *testing.T) {
	q := "type:File foo bar service:dropbox before:2017-01-1 service:invalid service:Google x"
	reg := test.DefaultRegistry()
	reg.RegisterAccountType(cloudsearch.Dropbox, nil, nil)
	reg.RegisterAccountType(cloudsearch.Google, nil, nil)

	parsed := cloudsearch.ParseQuery(q, "1", reg)
	if parsed.RawText != q {
		t.Fatal()
	}

	if *parsed.Before != time.Date(2017, 1, 1, 0, 0, 0, 0, time.UTC) {
		t.Fatal(parsed)
	}

	if parsed.After != nil {
		t.Fatal(parsed)
	}

	if !reflect.DeepEqual(parsed.AccountTypes, []cloudsearch.AccountType{cloudsearch.Dropbox, cloudsearch.Google}) {
		t.Fatal(parsed)
	}

	if !reflect.DeepEqual(parsed.ContentTypes, []cloudsearch.ContentType{cloudsearch.File}) {
		t.Fatal(parsed)
	}

	if parsed.Text != "foo bar     x" {
		t.Fatal(parsed.Text)
	}

}
