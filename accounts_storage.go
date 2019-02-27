package cloudsearch

import (
	"fmt"
	"time"
)

type AccountsStorage interface {
	Close()
	All() ([]AccountData, error)
	Active() ([]AccountData, error)
	Save(*AccountData) error
	Delete(accountId string) error
}

type AccountData struct {
	ID           string `storm:"id"`
	ExternalId   string
	Token        string
	Expiry       time.Time
	RefreshToken string
	TokenType    string
	AccountType  AccountType
	Name         string
	Email        string
	Active       bool
	Description  string
}

func (a *AccountData) String() string {
	return fmt.Sprintf("%s_%s", a.ID, a.AccountType)
}

func (a *AccountData) ShouldReauth() bool {
	return !a.Expiry.IsZero() && a.Expiry.Before(time.Now().Add(time.Minute*30))
}
