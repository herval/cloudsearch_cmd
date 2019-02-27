package google

import (
	"github.com/herval/cloudsearch"
	"context"
	"fmt"
	"golang.org/x/oauth2"
	goauth2 "google.golang.org/api/oauth2/v2"
	"net/http"
)

func NewAuthenticator(
	authService cloudsearch.OAuth2Authenticator,
	accounts cloudsearch.AccountsStorage,
	localRedirectUrl string,
) cloudsearch.IdentityService {
	return &GoogleAuth{
		authService,
		accounts,
		localRedirectUrl,
	}
}

type GoogleAuth struct {
	oauth2           cloudsearch.OAuth2Authenticator
	accounts         cloudsearch.AccountsStorage
	localRedirectUrl string
}

func (g *GoogleAuth) RefreshAccountIfNeeded(a cloudsearch.AccountData) (acc cloudsearch.AccountData, accountChanged bool, err error) {
	account, shouldSave, err := g.oauth2.RefreshTokenIfNeeded(a, g.localRedirectUrl)
	if err != nil {
		return a, false, err
	}
	if shouldSave {
		err = g.accounts.Save(&account)
		if err != nil {
			return a, false, err
		}
	}

	return account, shouldSave, err
}

func (g *GoogleAuth) FetchIdentityInfo(a cloudsearch.AccountData) (*cloudsearch.AccountData, error) {
	c := NewHttpClient(a)
	client, err := goauth2.New(c)
	if err != nil {
		return nil, err
	}

	res, err := client.Userinfo.V2.Me.Get().Do()
	if err != nil {
		return nil, err
	}

	a.ExternalId = res.Id
	a.Email = res.Email
	a.Name = res.Name
	a.Description = fmt.Sprintf("%s (%s)", res.Name, res.Email)

	return &a, nil
}

func NewHttpClient(a cloudsearch.AccountData) *http.Client {
	tok := &oauth2.Token{
		AccessToken:  a.Token,
		TokenType:    a.TokenType,
		RefreshToken: a.RefreshToken,
		Expiry:       a.Expiry,
	}
	src := oauth2.StaticTokenSource(tok)
	cli := oauth2.NewClient(context.Background(), src)

	return cli
}
