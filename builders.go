package cloudsearch

type AuthBuilder func(accountType AccountType) (IdentityService, error)

type SearchableBuilder func(account AccountData) (fetchFns []SearchFunc, ids []string, err error)

