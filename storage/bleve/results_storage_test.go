package bleve_test

import (
	"github.com/herval/cloudsearch"
	"github.com/herval/cloudsearch/storage/bleve"
	"testing"
)

func searchable(t *testing.T) *bleve.BleveResultStorage {
	index, err := bleve.NewIndex("./", "")
	if err != nil {
		t.Fatal(err)
	}

	s := bleve.NewBleveResultStorage(index).(*bleve.BleveResultStorage)

	err = s.Truncate()
	//if err != nil {
	//	t.Fatal("Couldnt truncate ", err)
	//}

	return s
}

func assertSave(r cloudsearch.Result, s cloudsearch.ResultsStorage, t *testing.T) *cloudsearch.Result {
	r, err := s.Save(r)
	if err != nil {
		t.Fatal("should save the result ", err, r)
	}
	return &r
}

func TestContentTypeQuery(t *testing.T) {
	s := searchable(t)
	defer s.Close()
	assertSave(
		cloudsearch.Result{
			ContentType: cloudsearch.Image,
			OriginalId:  "1",
		},
		s, t,
	)

	assertSave(
		cloudsearch.Result{
			ContentType: cloudsearch.Contact,
			OriginalId:  "2",
		},
		s, t,
	)

	q := cloudsearch.ParseQuery("type:image", "")
	if res, err := s.Search(q); err != nil || len(res) != 1 {
		t.Fatal("should find the image content only: ", res)
	}

	q = cloudsearch.ParseQuery("type:file", "")
	if res, err := s.Search(q); err != nil || len(res) != 0 {
		t.Fatal("should find no content!")
	}

}
