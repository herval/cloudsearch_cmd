package cloudsearch

import "context"

type SearchFunc func(query Query, context context.Context) <-chan Result
