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
	Url          string
}

func (a *AccountData) String() string {
	return fmt.Sprintf("%s_%s", a.ID, a.AccountType)
}

func (a *AccountData) ShouldReauth() bool {
	return !a.Expiry.IsZero() && a.Expiry.Before(time.Now().Add(time.Minute*30))
}

func (a *AccountData) JsonFields() map[string]interface{} {
	return map[string]interface{}{
		"id":          a.ID,
		"type":        a.AccountType,
		"name":        a.Name,
		"active":      a.Active,
		"email":       a.Email,
		"description": a.Description,
		"url":         a.Url,
	}
}
