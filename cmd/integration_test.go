package main_test

import (
	"github.com/herval/cloudsearch"
	"github.com/herval/cloudsearch/cmd"
	"context"
	"fmt"
	"testing"
)

func TestUncachedSearch(t *testing.T) {
	conf := main.NewConfig(cloudsearch.Env{"localhost", ":65432", "../"}, false)
	search := conf.SearchEngine

	t.Log("Searching...")

	q := cloudsearch.ParseQuery("foo", "123")
	res := search.Search(
		q,
		context.TODO(),
	)

	for r := range res {
		r.SetId() // not saved on db, so ID will be blank
		t.Log(fmt.Sprintf("%+v", r))
	}
}
