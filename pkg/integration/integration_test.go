package integration_test

import (
	"context"
	"fmt"
	"github.com/herval/cloudsearch/pkg"
	"github.com/herval/cloudsearch/pkg/config"
    "github.com/herval/cloudsearch/pkg/test"
    "testing"
)

func TestUncachedSearch(t *testing.T) {
	conf, err := config.NewConfig(cloudsearch.Env{"localhost", ":65432", "../"}, false)
	if err != nil {
		t.Fatal(err)
	}
	search := conf.SearchEngine

	t.Log("Searching...")

	q := cloudsearch.ParseQuery("foo", "123", test.DefaultRegistry())
	res := search.Search(
		q,
		context.TODO(),
	)

	for r := range res {
		r.SetId()
		t.Log(fmt.Sprintf("%+v", r))
	}
}
