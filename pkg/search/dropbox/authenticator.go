package dropbox

import (
	"github.com/herval/cloudsearch/pkg"
	"github.com/herval/dropbox-sdk-go-unofficial/dropbox"
	"github.com/herval/dropbox-sdk-go-unofficial/dropbox/users"
)

func NewAuthenticator() cloudsearch.IdentityService {
	return &DropboxAuth{
	}
}

type DropboxAuth struct {
}

func (d *DropboxAuth) RefreshAccountIfNeeded(a cloudsearch.AccountData) (acc cloudsearch.AccountData, accountChanged bool, err error) {
	return a, false, nil
}

func (d *DropboxAuth) FetchIdentityInfo(data cloudsearch.AccountData) (*cloudsearch.AccountData, error) {
	c := dropbox.Config{
		Token:    data.Token,
		LogLevel: dropbox.LogOff,
	}
	acc, err := users.New(c).GetCurrentAccount()
	if err != nil {
		return nil, err
	}

	data.Name = acc.Name.DisplayName
	data.ExternalId = acc.AccountId
	data.Email = acc.Email
	data.Description = acc.Email

	return &data, nil
}
