package cloudsearch


type Config struct {
	Env              Env
	AccountsStorage  AccountsStorage
	SearchEngine     *SearchEngine
	AuthBuilder      AuthBuilder
	ResultsStorage   ResultsStorage
	AuthService      OAuth2Authenticator
}