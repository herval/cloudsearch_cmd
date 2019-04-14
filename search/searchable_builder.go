package search

import (
	"fmt"
	"github.com/herval/cloudsearch"
	"github.com/herval/cloudsearch/search/bleve"
	"github.com/herval/cloudsearch/search/dropbox"
	"github.com/herval/cloudsearch/search/google"
	"github.com/pkg/errors"
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

func NewRemoteSearchablesBuilder(
	authBuilder cloudsearch.AuthBuilder,
) cloudsearch.SearchableBuilder {
	return func(account cloudsearch.AccountData) ([]cloudsearch.SearchFunc, []string, error) {
		switch account.AccountType {
		case cloudsearch.Dropbox:
			return []cloudsearch.SearchFunc{dropbox.NewSearch(account)},
				[]string{"dropbox"},
				nil

		case cloudsearch.Google:
			return google.SearchablesFor(
				google.NewHttpClient(account),
				account,
			)

		default:
			return nil, nil, errors.New(fmt.Sprintf("Cannot search for type: %s", string(account.AccountType)))
		}
	}
}
