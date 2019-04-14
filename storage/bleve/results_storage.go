package bleve

import (
	"encoding/json"
	bl "github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/document"
	m "github.com/blevesearch/bleve/mapping"
	"github.com/blevesearch/bleve/search"
	"github.com/blevesearch/bleve/search/query"
	"github.com/herval/cloudsearch"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"strings"
	"time"
)

type BleveResultStorage struct {
	index   bl.Index
	mapping m.IndexMapping
}

// store a Result in a search-optimized way
type searchableResult struct {
	Body        string
	Title       string
	Permalink   string
	Labels      string // space-separated list of labels
	Timestamp   time.Time
	Type        string
	ContentType string
	AccountType string
	Favorited   bool
	AccountId   string

	OriginalData string // a serializable json version of the Result
}

func NewBleveResultStorage(index bl.Index) cloudsearch.ResultsStorage {
	//logrus.SetLevel(cloudsearch.LogLevel)
	docs, err := index.DocCount()
	if err == nil {
		logrus.Debug("Opening index w/ docs: ", docs)
	}

	return &BleveResultStorage{
		index:   index,
		mapping: index.Mapping(),
	}
}

func (f *BleveResultStorage) DeleteAllFromAccount(accountId string) error {
	res, err := f.findIds(
		match(accountId, "AccountId", 1.0),
	)
	if err != nil {
		return err
	}

	for r := range res {
		err = f.index.Delete(r)
		if err != nil {
			logrus.Error("Deleting ", err)
		}
	}

	return nil
}

func (f *BleveResultStorage) AllFavoritedIds() ([]string, error) {
	res, err := f.findIds(
		matchBool(true, "Favorited", 1.0),
	)
	if err != nil {
		return nil, err
	}

	var r []string
	for i := range res {
		r = append(r, i)
	}

	return r, nil
}

func (s *BleveResultStorage) FindOlderThan(maxTime time.Time) (<-chan cloudsearch.Result, error) {
	var zero time.Time

	d, err := s.findIds(
		timeRange("Timestamp", &zero, &maxTime),
	)

	res := make(chan cloudsearch.Result)
	go func() {
		defer close(res)

		for id := range d {
			r, err := s.Get(id)
			if err != nil {
				logrus.Error("fetching res ", err)
				continue
			}
			if r != nil {
				res <- *r
			}
		}
	}()

	return res, err
}

func (s *BleveResultStorage) Delete(resultId string) error {
	err := s.index.Delete(resultId)
	if err != nil {
		return err
	}
	return nil
}

func (s *BleveResultStorage) findIds(q query.Query) (<-chan string, error) {
	res := make(chan string)

	go func() {
		defer close(res)

		from := 0
		size := 20
		for {
			req := bl.NewSearchRequestOptions(q, size, from, false)
			r, err := s.index.Search(req)
			if err != nil {
				logrus.Error("searching for ids ", err)
				break
			}
			from += size

			if len(r.Hits) == 0 {
				break
			}

			for _, r := range r.Hits {
				res <- r.ID
			}

		}
	}()

	return res, nil
}

func (s *BleveResultStorage) Close() {
	_  = s.index.Close()
}

func (s *BleveResultStorage) Get(resultId string) (*cloudsearch.Result, error) {
	doc, err := s.index.Document(resultId)
	if err != nil {
		return nil, err
	}

	res, err := toResult(doc, 0)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *BleveResultStorage) Merge(r cloudsearch.Result) (cloudsearch.Result, error) {
	r.SetId()
	//existing, err := s.Get(r.Id)
	//if err != nil {
	//	logrus.Error("Couldn't get from cache:", err)
	//}

	return s.Save(r)
}

func (s *BleveResultStorage) Save(result cloudsearch.Result) (cloudsearch.Result, error) {
	if result.OriginalId == "" {
		return result, errors.New("original id must be set")
	}

	result.SetId()

	if result.Status == 0 {
		result.Status = cloudsearch.ResultFound
	}

	if len(result.Id) != 32 {
		return result, errors.New("Id must be MD5 encoded!")
	}

	if result.Timestamp.IsZero() && result.ContentType != cloudsearch.Contact {
		logrus.Debug("TIMESTAMP IS NOT SET: ", result)
		//result.Timestamp = time.Now()
	}
	result.CachedAt = time.Now()

	// second precision is enough
	result.Timestamp = result.Timestamp.Truncate(time.Second)
	result.CachedAt = result.CachedAt.Truncate(time.Second)

	return result, s.index.Index(result.Id, searchable(result))
}

func (f *BleveResultStorage) IsFavorite(resultId string) (bool, error) {
	res, err := f.Get(resultId)
	if err != nil {
		logrus.Error("Error finding fave: ", err)
		return false, errors.Wrap(err, "finding favorite")
	}

	return res != nil && res.Favorited, nil
}

func (f *BleveResultStorage) ToggleFavorite(resultId string) (bool, error) {
	res, err := f.Get(resultId)
	if err != nil {
		return false, err
	}

	if res == nil {
		logrus.Debug("Trying to toggle result not found: ", resultId)
		return false, nil
	}

	res.Favorited = !res.Favorited
	_, err = f.Save(*res)

	return res.Favorited, err
}

func (f *BleveResultStorage) AllFavorited() ([]cloudsearch.Result, error) {
	ids, err := f.AllFavoritedIds()
	if err != nil {
		return nil, err
	}

	res := []cloudsearch.Result{}
	for _, d := range ids {
		r, err := f.Get(d)
		if err != nil {
			logrus.Error("getting ", d, ": ", err)
			continue
		}
		res = append(res, *r)
	}

	return res, nil
}

func searchable(result cloudsearch.Result) searchableResult {
	d, _ := json.Marshal(result)
	return searchableResult{
		Body:         result.Body,
		Permalink:    result.Permalink,
		Title:        result.Title,
		Timestamp:    result.Timestamp,
		OriginalData: string(d),
		Type:         "searchableResult",
		AccountId:    result.AccountId,
		AccountType:  string(result.AccountType),
		ContentType:  string(result.ContentType),
	}
}

// finds a single page of results
func (s *BleveResultStorage) find(q query.Query) ([]cloudsearch.Result, error) {
	req := bl.NewSearchRequestOptions(q, 20, 0, false)

	res, err := s.index.Search(req)
	if err != nil {
		return nil, err
	}
	hits := res.Hits

	return toResults(s.index, hits)
}

func (s *BleveResultStorage) Search(q cloudsearch.Query) ([]cloudsearch.Result, error) {
	logrus.Debug("Searching Cache: ", q)

	subqueries := []query.Query{
		anyOf(matchTypes(cloudsearch.ContentTypeStrings(q.ContentTypes), "ContentType")...),  // match any content type provided
		anyOf(matchTypes(cloudsearch.AccountTypesStrings(q.AccountTypes), "AccountType")...), // match any account type provided
		timeRange("Timestamp", q.Before, q.After),
	}

	// empty searches for content types may still yield results
	str := strings.Trim(q.Text, " ")
	if str != "" {
		subqueries = append(subqueries,
			anyOf( // any match on body, title, permalink
				match(q.Text, "Title", 3),
				match(q.Text, "Body", 3),
				prefix(q.Text, "Title", 2),
				prefix(q.Text, "Body", 2),
				prefix(q.Text, "Permalink", 1),
				fuzzy(q.Text, "Title", 1.5),
				fuzzy(q.Text, "Body", 1.5),
			),
		)
	}

	if str == "" && len(q.ContentTypes) == 0 {
		return nil, errors.New("Cannot search - empty query")
	}

	union := allOf(subqueries...)

	// TODO increase the score for newer content
	// TODO time ranges
	// TODO accounts

	return s.find(union)
}

func (f *BleveResultStorage) Truncate() error {
	ids, err := f.findIds(allOf(query.NewMatchAllQuery()))
	if err != nil {
		logrus.Error("Could not truncate: ", err)
		return err
	}

	for i := range ids {
		err = f.Delete(i)
		if err != nil {
			return err
		}
	}

	return nil
}

func toResults(index bl.Index, hits search.DocumentMatchCollection) ([]cloudsearch.Result, error) {
	res := []cloudsearch.Result{}
	for _, h := range hits {
		d, err := index.Document(h.ID)
		if err != nil {
			return nil, err
		}

		// reconstruct the documents based on the hits
		r, err := toResult(d, h.Score)
		if err != nil {
			logrus.Error("Invalid data on " + d.ID + ": " + err.Error())
			// TODO clean up the record?
		}
		if r != nil {
			res = append(res, *r)
		}
	}

	return res, nil
}

func timeRange(field string, before *time.Time, after *time.Time) query.Query {
	if before == nil && after == nil {
		return nil
	}

	if before == nil {
		t := time.Now()
		before = &t
	}
	if after == nil {
		t := time.Unix(0, 0)
		after = &t
	}

	y := true // wat
	q := bl.NewDateRangeInclusiveQuery(*after, *before, &y, &y)
	q.SetField(field)
	return q
}

func allOf(queries ...query.Query) query.Query {
	res := []query.Query{}
	for _, q := range queries {
		if q != nil {
			res = append(res, q)
		}
	}
	return bl.NewConjunctionQuery(res...)
}

func anyOf(queries ...query.Query) query.Query {
	if len(queries) == 0 {
		return nil
	}

	return bl.NewDisjunctionQuery(queries...)
}

func matchTypes(contentTypes []string, field string) []query.Query {
	res := []query.Query{}

	for _, c := range contentTypes {
		q := bl.NewMatchQuery(c)
		q.SetField(field)
		res = append(res, q)
	}

	return res
}

func match(query string, field string, boost float64) query.Query {
	mm := bl.NewMatchQuery(query)
	mm.SetField(field)
	mm.SetBoost(boost)
	return mm
}

func matchBool(boolean bool, field string, boost float64) query.Query {
	mm := bl.NewBoolFieldQuery(boolean)
	mm.SetField(field)
	mm.SetBoost(boost)
	return mm
}

func prefix(query string, field string, boost float64) query.Query {
	prefixMatch := bl.NewPrefixQuery(query)
	prefixMatch.SetField(field)
	prefixMatch.SetBoost(boost)
	return prefixMatch
}

func fuzzy(query string, field string, boost float64) *query.FuzzyQuery {
	fuzzyMatch := bl.NewFuzzyQuery(query)
	fuzzyMatch.SetFuzziness(1)
	fuzzyMatch.SetBoost(boost)
	fuzzyMatch.SetField(field)
	return fuzzyMatch
}

type bleveResult struct {
	ID     string                 `json:"id"`
	Fields map[string]interface{} `json:"fields"`
}

func toResult(doc *document.Document, hitScore float64) (*cloudsearch.Result, error) {
	if doc == nil {
		return nil, nil
	}

	// first map it all to a bleve result thing
	rv := bleveResult{
		ID:     doc.ID,
		Fields: map[string]interface{}{},
	}
	for _, field := range doc.Fields {
		var newval interface{}
		switch field := field.(type) {
		case *document.TextField:
			newval = string(field.Value())
			// no need for any of this w/ current impl
			//case *document.NumericField:
			//	n, err := field.Number()
			//	if err == nil {
			//		newval = n
			//	}
			//case *document.BooleanField:
			//	d, err := field.Boolean()
			//	if err == nil {
			//		newval = d
			//	}
			//case *document.DateTimeField:
			//	d, err := field.DateTime()
			//	if err == nil {
			//		newval = d.Format(time.RFC3339Nano)
			//	}
		}

		existing, existed := rv.Fields[field.Name()]
		if existed {
			switch existing := existing.(type) {
			case []interface{}:
				rv.Fields[field.Name()] = append(existing, newval)
			case interface{}:
				arr := make([]interface{}, 2)
				arr[0] = existing
				arr[1] = newval
				rv.Fields[field.Name()] = arr
			}
		} else {
			//fmt.Println(field, newval)
			rv.Fields[field.Name()] = newval
		}
	}

	// then map all fields
	res := cloudsearch.Result{}
	err := json.Unmarshal([]byte(rv.Fields["OriginalData"].(string)), &res)
	if err != nil {
		return nil, err
	}
	res.CacheHitScore = hitScore

	return &res, nil
}
