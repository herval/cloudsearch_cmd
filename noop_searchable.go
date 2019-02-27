package cloudsearch

import "context"

func NoopSearchable() SearchFunc {
	return func(query Query, context context.Context) <-chan Result {
		res := make(chan Result)
		defer close(res)

		return res
	}
}
