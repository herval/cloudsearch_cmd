package dropbox

import "github.com/herval/cloudsearch"

func AuthBuilder(accountType cloudsearch.AccountType) (cloudsearch.IdentityService, error) {
	return NewAuthenticator(), nil
}

func SearchBuilder(account cloudsearch.AccountData) (fetchFns []cloudsearch.SearchFunc, ids []string, err error) {
	return []cloudsearch.SearchFunc{NewSearch(account)},
		[]string{"dropbox"},
		nil
}
