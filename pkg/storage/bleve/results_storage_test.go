package bleve_test

import (
	"github.com/herval/cloudsearch/pkg"
	"github.com/herval/cloudsearch/pkg/storagerage/bleve"
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

func TestFavorite(t *testing.T) {
	s := searchable(t)
	defer s.Close()
	r := assertSave(
		cloudsearch.Result{
			Favorited:   false,
			ContentType: cloudsearch.Message,
			OriginalId:  "3",
		},
		s, t,
	)

	if r, err := s.Get(r.Id); err != nil || r == nil || r.Favorited {
		t.Fatal("Should find the unfavorited result ", err, r)
	}

	if fav, err := s.ToggleFavorite(r.Id); err != nil || !fav {
		t.Fatal("should toggle favorite ", err, fav)
	}

	if fav, err := s.IsFavorite(r.Id); err != nil || !fav {
		t.Fatal("should be faved ", err, fav)
	}

	if fav, err := s.AllFavorited(); err != nil || len(fav) != 1 || !fav[0].Favorited {
		t.Fatal("should be faved ", err, fav)
	}
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
