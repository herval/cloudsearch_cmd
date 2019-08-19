package integration_test

import (
	"context"
	"fmt"
	"github.com/herval/cloudsearch/pkg"
	"github.com/herval/cloudsearch/cmd"
	"testing"
)

func TestUncachedSearch(t *testing.T) {
	conf := main.NewConfig(cloudsearch.Env{"localhost", ":65432", "../"}, false)
	search := conf.SearchEngine

	t.Log("Searching...")

	q := cloudsearch.ParseQuery("foo", "123", cloudsearch.NewRegistry())
	res := search.Search(
		q,
		context.TODO(),
	)

	for r := range res {
		r.SetId()
		t.Log(fmt.Sprintf("%+v", r))
	}
}
