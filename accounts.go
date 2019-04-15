package cloudsearch

import "github.com/pkg/errors"

type AccountType string

const (
	Dropbox AccountType = "Dropbox"
	Google  AccountType = "Google"
)

var SupportedAccountTypes = []AccountType{
	Dropbox, Google,
}

var SupportedAccountTypesStr = AccountTypesStrings(SupportedAccountTypes)

func RegisterAccountTypes(acc ...AccountType) {
	for _, c := range acc {
		SupportedAccountTypes = append(SupportedAccountTypes, c)
		SupportedAccountTypesStr = append(SupportedAccountTypesStr, string(c))
	}
}

func AccountTypeIncluded(list []AccountType, a AccountType) bool {
	for _, r := range list {
		if r == a {
			return true
		}
	}
	return false
}

func AccountTypesStrings(c []AccountType) []string {
	var res []string
	for _, cc := range c {
		res = append(res, string(cc))
	}
	return res
}

func ParseAccountType(str string) (AccountType, error) {
	for _, s := range SupportedAccountTypesStr {
		if s == str {
			return AccountType(s), nil
		}
	}
	return "", errors.New("Unsupported type: " + str)
}
