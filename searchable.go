package cloudsearch

import "context"

type SearchFunc func(query Query, context context.Context) <-chan Result

type SyncSearchFunc func(query Query, context context.Context) ([]Result, error)

func NewAsyncSearchable(
	search SyncSearchFunc,
) SearchFunc {
	return func(query Query, context context.Context) <-chan Result {
		out := make(chan Result)
		go func() {
			defer close(out)

			res, err := search(query, context)
			if res != nil {
				for _, r := range res {
					out <- r
				}
			}

			if err != nil {
				out <- Result{
					Title:  err.Error(),
					Status: ResultError,
				}
			}
		}()

		return out
	}
}
