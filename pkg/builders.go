package cloudsearch

type AuthBuilder func(accountType AccountType) (IdentityService, error)

// build a set of search functions w/ individual search ids for a given account
type SearchableBuilder func(account AccountData) (fetchFns []SearchFunc, ids []string, err error)

