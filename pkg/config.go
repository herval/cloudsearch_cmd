package cloudsearch

type Config struct {
	Env             Env
	AccountsStorage AccountsStorage
	SearchEngine    *SearchEngine
	ResultsStorage  ResultsStorage
	AuthService     OAuth2Authenticator
	Registry        *Registry
}
