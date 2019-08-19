package bleve

import (
	"github.com/herval/cloudsearch/pkg"
	"context"
)

func NewIndexedResultsSearchable(storage cloudsearch.ResultsStorage) cloudsearch.SearchFunc {
	return func(query cloudsearch.Query, context context.Context) <-chan cloudsearch.Result {
		res := make(chan cloudsearch.Result)

		go func() {
			defer close(res)

			d, err := storage.Search(query)
			if err != nil && context.Err() != nil {
				res <- cloudsearch.Result{
					Title:  err.Error(),
					Status: cloudsearch.ResultError,
				}
				return
			}

			for _, dd := range d {
				res <- dd
			}
		}()

		return res
	}
}