package cloudsearch

import (
	"github.com/pkg/errors"
)

// keep a centralized mapping between account types and search/auth builders,
// for easy extension for new services
type Registry struct {
	accountTypes    []AccountType
	accountTypeStrs []string
	searchables     map[AccountType]SearchableBuilder
	authorizers     map[AccountType]AuthBuilder
}

func NewRegistry() *Registry {
	return &Registry{
		accountTypes:    []AccountType{},
		accountTypeStrs: []string{},
		searchables:     map[AccountType]SearchableBuilder{},
		authorizers:     map[AccountType]AuthBuilder{},
	}
}

func (r *Registry) RegisterAccountType(
	acc AccountType,
	searchBuilder SearchableBuilder,
	authBuilder AuthBuilder,
) {
	r.accountTypes = append(r.accountTypes, acc)
	r.accountTypeStrs = append(r.accountTypeStrs, string(acc))
	r.searchables[acc] = searchBuilder
	r.authorizers[acc] = authBuilder
}

func (r *Registry) SupportedAccountTypes() []AccountType {
	return r.accountTypes
}

func (r *Registry) SearchBuilder(account AccountData) (fetchFns []SearchFunc, ids []string, err error) {
	b, ok := r.searchables[account.AccountType]
	if !ok {
		return nil, nil, errors.New("No search builder found for type: " + string(account.AccountType))
	}

	return b(account)
}

func (r *Registry) AuthBuilder(accountType AccountType) (IdentityService, error) {
	b, ok := r.authorizers[accountType]
	if !ok {
		return nil, errors.New("No auth builder found for type: " + string(accountType))
	}

	return b(accountType)
}

func (r *Registry) IsAccountTypeSupported(accountType AccountType) bool {
	_, ok := r.searchables[accountType]
	return ok
}

func (r *Registry) ParseAccountType(str string) (AccountType, error) {
	for _, s := range r.accountTypes {
		if string(s) == str {
			return s, nil
		}
	}
	return "", errors.New("Unsupported type: " + str)
}
func (r *Registry) SupportedAccountTypesStr() []string {
	return r.accountTypeStrs
}
