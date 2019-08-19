package cloudsearch

import (
    "github.com/pkg/errors"
)

// keep a centralized mapping between account types and search/auth builders,
// for easy extension for new services
type Registry struct {
    searchables map[AccountType]SearchableBuilder
    authorizers map[AccountType]AuthBuilder

    // using a map to keep them unique
    accountTypes map[AccountType]interface{}
    contentTypes map[ContentType]interface{}
}

func NewRegistry() *Registry {
    return &Registry{
        accountTypes: map[AccountType]interface{}{},
        searchables:  map[AccountType]SearchableBuilder{},
        authorizers:  map[AccountType]AuthBuilder{},
        contentTypes: map[ContentType]interface{}{},
    }
}

func (r *Registry) RegisterContentTypes(ct ...ContentType) {
    for _, c := range ct {
        r.contentTypes[c] = nil
    }
}

func (r *Registry) RegisterAccountType(
    acc AccountType,
    searchBuilder SearchableBuilder,
    authBuilder AuthBuilder,
) {
    r.accountTypes[acc] = nil
    r.searchables[acc] = searchBuilder
    r.authorizers[acc] = authBuilder
}

func (r *Registry) SupportedAccountTypes() []AccountType {
    var res []AccountType
    for k, _ := range r.accountTypes {
        res = append(res, k)
    }
    return res
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
    for s, _ := range r.accountTypes {
        if string(s) == str {
            return s, nil
        }
    }
    return "", errors.New("Unsupported type: " + str)
}

func (r *Registry) SupportedAccountTypesStr() []string {
    var res []string
    for r, _ := range r.accountTypes {
        res = append(res, string(r))
    }
    return res
}

func (r *Registry) supportedContentTypesStr() []string {
    var res []string
    for r, _ := range r.contentTypes {
        res = append(res, string(r))
    }
    return res
}
