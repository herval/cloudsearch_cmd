package mock

import (
	"context"
	"time"

	"github.com/herval/cloudsearch/pkg"
)

type Searchable struct {
}

func (s *Searchable) SearchSnippets(query cloudsearch.Query, context context.Context) <-chan cloudsearch.Result {
	res := make(chan cloudsearch.Result)

	go func() {
		res <- cloudsearch.Result{
			Id:    "1",
			Title: "found.gif",
		}
		res <- cloudsearch.Result{
			Id:    "2",
			Title: "found2.gif",
		}

		time.Sleep(2 * time.Second)
		close(res)
	}()

	return res
}
