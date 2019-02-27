package cloudsearch

import (
	"context"
	"github.com/sirupsen/logrus"
	"sync"
)

func NewCachedSearchable(
	name string,
	results ResultsStorage,
	local SearchFunc,
	remote SearchFunc,
	accountType AccountType,
) SearchFunc {
	return func(query Query, ctx context.Context) <-chan Result {
		m := NewStopwatch("cached_and_remote_search_" + name)
		res := make(chan Result)
		var wg sync.WaitGroup

		// search on results
		if query.SearchMode == All || query.SearchMode == Cache {
			logrus.Debug("Searching local " + name)
			wg.Add(1)
			go func(res chan Result) {
				n := "cached_search_" + name
				m := NewStopwatch(n)
				i := 0
				rr := local(query, ctx)
				for r := range rr {
					if ctx.Err() != nil {
						break
					}
					res <- r
					i += 1
				}
				// TODO go through cachedResult objects and check if they still exist if they're stale
				wg.Done()
				m.Lap()
				logrus.Debug(n, " results: ", i)
			}(res)
		}

		// search underlying service
		if query.SearchMode == All || query.SearchMode == Live {
			logrus.Debug("Searching remote " + name)
			wg.Add(1)
			go func(res chan Result) {
				i := 0
				n := "remote_search_" + name
				m := NewStopwatch(n)
				rs := remote(query, ctx)
				for r := range rs {
					if ctx.Err() != nil {
						break
					}

					switch r.Status {
					case ResultFound:
						// save found results on results
						var err error
						r, err = results.Merge(r)
						if err != nil {
							logrus.Error("Couldn't merge result:", err)
						}

					case ResultNotFound:
						err := results.Delete(r.Id)
						logrus.Debug("Deleting removed content: ", r)
						if err != nil {
							logrus.Error("Couldn't remove cached result:", err)
						}
					}
					res <- r
					i += 1
				}
				wg.Done()
				m.Lap()
				logrus.Debug(n, " results: ", i)
			}(res)
		}

		// wait for all routines to finish before closing the results stream
		go func(res chan Result) {
			wg.Wait()
			m.Lap()
			close(res)
		}(res)

		return res
	}
}
