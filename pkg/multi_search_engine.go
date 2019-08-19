package cloudsearch

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func NewMultiSearch(
	env Env,
	accounts AccountsStorage,
	results ResultsStorage,
	registry *Registry,
	filterBuilder func(q Query) []ResultFilter,
) SearchEngine {
	a := SearchEngine{
		env:                env,
		accounts:           accounts,
		currentSearchables: []SearchFunc{},
		FilterBuilder:      filterBuilder,
		registry:           registry,
		results:            results,
	}

	err := a.Refresh()
	if err != nil {
		logrus.Error("Setting up search: ", err)
	}

	return a
}

// search multiple searchables and multiplex the results into a single channel
type SearchEngine struct {
	lock               sync.Mutex
	env                Env
	currentSearchables []SearchFunc
	accounts           AccountsStorage
	results            ResultsStorage
	FilterBuilder      func(q Query) []ResultFilter
	registry           *Registry

	// allow building composable searchables (eg support caching and filtering). One account can have multiple searchables.
}

// rebuild the searchables list, closing previously open searchables
func (s *SearchEngine) Refresh() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	a, err := s.accounts.All()
	if err != nil {
		return err
	}

	s.currentSearchables = []SearchFunc{}

	for _, acc := range a {
		if acc.ShouldReauth() {
			auth, err := s.registry.AuthBuilder(acc.AccountType)
			if err != nil {
				return errors.Wrap(err, "Could not build authenticator")
			}

			acc, _, err = auth.RefreshAccountIfNeeded(acc)
			if err != nil {
				logrus.Error("Could not search " + acc.Description + ": " + err.Error())
				//return errors.Wrap(err, "Could not refresh account")
				continue
			}
			// TODO mark failed as inactive?
		}

		searchables, _, err := s.registry.SearchBuilder(acc)
		if err != nil {
			return err
		}
		s.currentSearchables = append(s.currentSearchables, searchables...)
	}

	return nil
}

func (s *SearchEngine) Search(query Query, ctx context.Context) <-chan Result {
	ctx, _ = context.WithTimeout(ctx, time.Second*15) // dont wait too long for downstream answers - some are pretty pretty slow

	m := NewStopwatch("multisearch_" + query.SearchId)
	searchables := s.currentSearchables
	filters := s.FilterBuilder(query)

	results := make(chan Result)

	var finished int32 = 0
	done := make(chan bool, len(searchables))

	logrus.WithFields(map[string]interface{}{
		"sources": len(searchables),
		"queryId": query.SearchId,
		"query":   query.RawText,
	}).Info("Searching datasources")

	withFilters := func(d SearchFunc) {
		res := d(query, ctx)
		for d := range res {
			if ctx.Err() != nil { // cancel it all
				break
			}

			c := &d
			skip := false
			if c.Id == "" {
				c.SetId()
			}

			// apply filters
			for _, filterOut := range filters {
				if ctx.Err() != nil {
					skip = true
					break
				}

				c = filterOut(query, d)
				if c == nil {
					skip = true
					break
				}
			}

			if !skip && ctx.Err() == nil {
				results <- *c
			}
		}
		done <- true
	}

	if len(searchables) > 0 {
		for _, d := range searchables {
			go withFilters(d)
		}

		go func() {
			for finished < int32(len(searchables)) {
				select {
				case <-ctx.Done():
					logrus.Debug("Request cancelled, cancelling all searches")
					atomic.StoreInt32(&finished, int32(len(searchables)))
				case <-done:
					atomic.AddInt32(&finished, 1)
				}
			}

			logrus.Debug("Closing search " + query.SearchId)
			m.Lap()
			close(results)
		}()

	} else {
		// close right away
		defer close(results)
	}

	return results
}

func (s *SearchEngine) SaveAccount(data *AccountData) error {
	err := s.accounts.Save(data)
	if err != nil {
		return err
	}

	return s.Refresh()
}

func (s *SearchEngine) AllAccounts() ([]AccountData, error) {
	return s.accounts.All()
}

func (s *SearchEngine) DeleteAccount(id string) error {
	_, err := s.results.DeleteAllFromAccount(id)
	if err != nil {
		return err
	}

	err = s.accounts.Delete(id)
	if err != nil {
		return err
	}

	return s.Refresh()
}

// watch for expired tokens and try to refresh them in a loop
func (s *SearchEngine) WatchTokens() {
	for true {
		acc, err := s.accounts.All()
		if err != nil {
			logrus.Error("Getting accts", err)
			continue
		}

		refresh := false
		for _, a := range acc {
			if a.RefreshToken != "" {
				// TODO if auth can't be established fast, fail

				auth, err := s.registry.AuthBuilder(a.AccountType)
				if err != nil {
					logrus.Error("Auth building", err)
					continue
				}

				_, changed, err := auth.RefreshAccountIfNeeded(a)
				if err != nil {
					logrus.Error("Refreshing acc ", err)
					continue
				}

				if changed {
					logrus.Debug("Account auth changed, refreshing")
					refresh = true
				}
			}
		}

		if refresh {
			err = s.Refresh()
			if err != nil {
				logrus.Error("Ref", err)
			}
		}

		time.Sleep(time.Minute * 10)
	}
}
