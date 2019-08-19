package auth

import (
	authgateway2 "github.com/herval/authgateway"
	"github.com/herval/authgateway/client"
	"github.com/herval/cloudsearch/pkg"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

const DefaultGatewayUrl = "https://cloudsearch-auth.herokuapp.com"
const LocalGatewayUrl = "http://localhost:8080"

type AuthgatewayAuthenticator struct {
	client authgateway.AuthClient
}

func NewAuthenticator(client authgateway.AuthClient) cloudsearch.OAuth2Authenticator {
	return &AuthgatewayAuthenticator{
		client,
	}
}

func (a *AuthgatewayAuthenticator) AuthorizeUrl(acc cloudsearch.AccountType, redirectUrl string) (string, error) {
	return a.client.AuthorizeUrl(string(acc), redirectUrl)
}

func (a *AuthgatewayAuthenticator) AccountFromCode(acc cloudsearch.AccountType, code string, redirectUrl string) (*cloudsearch.AccountData, error) {
	tok, err := a.client.TokenFromCode(string(acc), code, redirectUrl)
	if err != nil {
		return nil, err
	}
	if tok == nil {
		return nil, errors.New("No token generated")
	}

	aa := a.AccountFor(acc, cloudsearch.AccountData{}, token(tok))
	return &aa, nil
}

func (a *AuthgatewayAuthenticator) AccountFor(accountType cloudsearch.AccountType, acc cloudsearch.AccountData, tok *oauth2.Token) cloudsearch.AccountData {
	acc.Active = true
	acc.Token = tok.AccessToken
	acc.Expiry = tok.Expiry
	acc.RefreshToken = tok.RefreshToken
	acc.TokenType = tok.TokenType
	acc.AccountType = accountType

	return acc
}

func (a *AuthgatewayAuthenticator) RefreshTokenIfNeeded(account cloudsearch.AccountData, redirectUrl string) (aa cloudsearch.AccountData, accountChanged bool, err error) {
	if !account.ShouldReauth() {
		// no need to update if not expired
		return account, false, nil
	}

	to, changed, err := a.client.RefreshToken(string(account.AccountType), tok(account), redirectUrl)
	if err != nil || !changed {
		return account, changed, err
	}

	acc := a.AccountFor(account.AccountType, account, token(&to))
	return acc, true, err
}

func tok(data cloudsearch.AccountData) authgateway2.Token {
	return authgateway2.Token{
		AccessToken:  data.Token,
		RefreshToken: data.RefreshToken,
		TokenType:    data.TokenType,
		Expiry:       data.Expiry,
	}
}

func token(token *authgateway2.Token) *oauth2.Token {
	return &oauth2.Token{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		Expiry:       token.Expiry,
	}
}
