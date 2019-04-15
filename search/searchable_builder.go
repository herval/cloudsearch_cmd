package search

import (
	"github.com/herval/cloudsearch"
	"github.com/herval/cloudsearch/search/bleve"
	"github.com/sirupsen/logrus"
)

func NewCachedSearchableBuilder(
	results cloudsearch.ResultsStorage,
	remotes cloudsearch.SearchableBuilder,
	cacheLocally bool,
) cloudsearch.SearchableBuilder {
	return func(a cloudsearch.AccountData) ([]cloudsearch.SearchFunc, []string, error) {
		search, ids, err := remotes(a)
		if err != nil {
			logrus.Error("Could not setup a remote searchable for ", a.ID, " - ", err)
			search = []cloudsearch.SearchFunc{
				cloudsearch.NoopSearchable(),
			}
		}

		if cacheLocally {
			// support cached results
			cached := []cloudsearch.SearchFunc{}
			for i, c := range search {
				cs := bleve.NewIndexedResultsSearchable(results)
				ca := cloudsearch.NewCachedSearchable(ids[i], results, cs, c, a.AccountType)
				cached = append(cached, ca)
			}
			search = cached
		}

		return search, ids, nil
	}
}
