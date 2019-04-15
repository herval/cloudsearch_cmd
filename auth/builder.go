package auth

import "github.com/herval/cloudsearch"

// a builder for the simple case: one auth config per account
func Builder(service cloudsearch.IdentityService) cloudsearch.AuthBuilder {
	return func(accountType cloudsearch.AccountType) (cloudsearch.IdentityService, error) {
		return service, nil
	}
}
