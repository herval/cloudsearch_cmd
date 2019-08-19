package google

import (
	"github.com/herval/cloudsearch/pkg"
)

func AuthBuilder(
	authService cloudsearch.OAuth2Authenticator,
	storage cloudsearch.AccountsStorage,
	googleOauthCallbackPath string,
) cloudsearch.AuthBuilder {
	return func(accountType cloudsearch.AccountType) (cloudsearch.IdentityService, error) {
		return NewAuthenticator(authService, storage, googleOauthCallbackPath), nil
	}
}

func SearchBuilder(account cloudsearch.AccountData) (fetchFns []cloudsearch.SearchFunc, ids []string, err error) {
	httpClient := NewHttpClient(account)
	drive, err := NewGoogleDrive(account, httpClient)
	if err != nil {
		return nil, nil, err
	}

	gmail, err := NewGmail(account, httpClient)
	if err != nil {
		return nil, nil, err
	}

	return []cloudsearch.SearchFunc{
		drive.SearchSnippets,
		gmail.SearchSnippets,
	}, []string{
		"drive",
		"gmail",
	}, nil
}
