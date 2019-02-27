package main

import (
	"errors"
	"github.com/herval/cloudsearch"
	"github.com/herval/cloudsearch/search/dropbox"
	"github.com/herval/cloudsearch/search/google"
)

func NewAuthBuilder(
	authService cloudsearch.OAuth2Authenticator,
	storage cloudsearch.AccountsStorage,
	googleOauthCallbackPath string,
) cloudsearch.AuthBuilder {
	return func(accountType cloudsearch.AccountType) (cloudsearch.IdentityService, error) {
		switch accountType {
		case cloudsearch.Google:
			return google.NewAuthenticator(authService, storage, googleOauthCallbackPath), nil

		case cloudsearch.Dropbox:
			return dropbox.NewAuthenticator(), nil

		default:
			return nil, errors.New("No Oauth2 provider for " + string(accountType))
		}
	}
}

