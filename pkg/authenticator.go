package cloudsearch

type IdentityService interface {
	FetchIdentityInfo(a AccountData) (*AccountData, error)
	RefreshAccountIfNeeded(a AccountData) (acc AccountData, accountChanged bool, err error)
}

type OAuth2Authenticator interface {
	AuthorizeUrl(acc AccountType, redirectUrl string) (string, error)
	AccountFromCode(acc AccountType, code string, redirectUrl string) (*AccountData, error)
	RefreshTokenIfNeeded(account AccountData, redirectUrl string) (a AccountData, accountChanged bool, err error)
}

type PasswordAuthenticator interface {
	AccountFromCredentials(username string, password string, server string) (*AccountData, error)
}
