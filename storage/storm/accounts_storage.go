package storm

import (
	"github.com/herval/cloudsearch"
	"errors"
	"fmt"
	"github.com/asdine/storm"
)

type AccountsStorage struct {
	s *storm.DB
}

func NewAccountsStorage(storagePath string) (cloudsearch.AccountsStorage, error) {
	db, err := storm.Open(cloudsearch.FileAt(storagePath, "accounts.db"))
	if err != nil {
		return nil, err
	}

	return &AccountsStorage{
		s: db,
	}, nil
}

func (s *AccountsStorage) All() ([]cloudsearch.AccountData, error) {
	res := make([]cloudsearch.AccountData, 0)
	err := s.s.All(&res)
	return res, err
}

func (s *AccountsStorage) Active() ([]cloudsearch.AccountData, error) {
	res := make([]cloudsearch.AccountData, 0)
	err := s.s.Find("Active", true, &res)
	return res, err
}

func (s *AccountsStorage) Delete(id string) error {
	return s.s.DeleteStruct(&cloudsearch.AccountData{ID: id})
}

func (s *AccountsStorage) Save(data *cloudsearch.AccountData) error {
	if data.ExternalId == "" || data.Name == "" {
		return errors.New("Invalid account: missing id information")
	}

	if data.ID == "" {
		// avoid re-registering same account by resetting the ID
		data.ID = cloudsearch.Md5(fmt.Sprintf("%s_%s", string(data.AccountType), data.ExternalId))
	}

	if len(data.ID) != 32 {
		return errors.New("Invalid id: must be an md5")
	}

	return s.s.Save(data)
}

func (s *AccountsStorage) Close() {
	_ = s.s.Close()
}
