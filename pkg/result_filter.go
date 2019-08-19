package cloudsearch

import (
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

type ResultFilter func(Query, Result) *Result

func Dedup(query Query) func(query Query, in Result) *Result {
	logrus.Debug("Setting up dedup for ", query)
	lock := sync.RWMutex{}
	alreadyPosted := map[string]bool{}

	return func(query Query, in Result) *Result {
		lock.RLock()
		posted := alreadyPosted[in.Id] == true
		lock.RUnlock()

		if !posted {
			lock.Lock()
			alreadyPosted[in.Id] = true
			lock.Unlock()

			return &in
		} else {
			logrus.Debug("Already posted, filtering:", in.Id)
			return nil
		}
	}
}

func FilterNotInRange(query Query, in Result) *Result {
	//logrus.Debug("Filtering not in range for ", query)
	minTime := time.Unix(0, 0)
	if query.After != nil {
		minTime = *query.After
	}

	maxTime := time.Now()
	if query.Before != nil {
		maxTime = *query.Before
	}

	if in.Timestamp.IsZero() || (in.Timestamp.After(minTime) && in.Timestamp.Before(maxTime)) {
		//logrus.Debug("OK: ", in.Timestamp, in.Title)
		return &in
	} else {
		logrus.Debug("Filtering out of range: ", in.Id)
		return nil
	}
}

func FilterContent(query Query, in Result) *Result {
	//logrus.Debug("Filtering content for ", in)
	types := query.ContentTypes

	if types == nil || len(types) == 0 ||
		typesInclude(types, in.ContentType) { // TODO files?
		return &in
	} else {
		logrus.Debug("Filtering by content type: ", in.Id)
		return nil
	}
}


func SetId(query Query, in Result) *Result {
	if in.Id == "" {
		in.SetId()
	}
	return &in
}

func typesInclude(types []ContentType, t ContentType) bool {
	for _, tt := range types {
		if t == tt {
			return true
		}
	}
	return false
}

//func LimitResults() ResultFilter {
//	return func(query Query, result Result) *Result {
//		max := query.MaxResults
//
//	go func() {
//		logrus.Debug("Filtering max results for ", query)
//		current := 0
//		for c := range in {
//			if current < max {
//				res <- c
//				current += 1
//			} else {
//				logrus.Debug("Filtering excess results: ", c.Id)
//			}
//		}
//
//		close(res)
//	}()
//
//	return res
//}
