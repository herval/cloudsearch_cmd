package cloudsearch

type AccountType string

const (
	Dropbox AccountType = "Dropbox"
	Google  AccountType = "Google"
)

//var SupportedAccountTypes = []AccountType{
//	Dropbox, Google,
//}
//
//var SupportedAccountTypesStr = AccountTypesStrings(SupportedAccountTypes)
//
//func RegisterAccountTypes(acc ...AccountType) {
//	for _, c := range acc {
//		SupportedAccountTypes = append(SupportedAccountTypes, c)
//		SupportedAccountTypesStr = append(SupportedAccountTypesStr, string(c))
//	}
//}

