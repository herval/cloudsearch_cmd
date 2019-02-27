package cloudsearch

import "time"

type ResultsStorage interface {
	Close()
	Save(result Result) (Result, error) // fully override a result
	Merge(result Result) (Result, error) // save a result, merging data such as favorite status if it's set
	Search(query Query) ([]Result, error)
	Get(resultId string) (*Result, error)

	FindOlderThan(maxTime time.Time) (<-chan Result, error)
	DeleteAllFromAccount(accountId string) error
	Delete(resultId string) error
}
