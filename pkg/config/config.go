package config

import (
	"net/http"
	"time"

	authgateway "github.com/herval/authgateway/client"
	"github.com/herval/cloudsearch/pkg"
	"github.com/herval/cloudsearch/pkg/auth"
	"github.com/herval/cloudsearch/pkg/search"
	"github.com/herval/cloudsearch/pkg/search/dropbox"
	"github.com/herval/cloudsearch/pkg/search/google"
	"github.com/herval/cloudsearch/pkg/storage/bleve"
	"github.com/herval/cloudsearch/pkg/storage/storm"
)

func NewConfig(env cloudsearch.Env, enableCaching bool) (cloudsearch.Config, error) {
	accounts, err := storm.NewAccountsStorage(env.StoragePath)
	if err != nil {
		return cloudsearch.Config{}, err
	}

	index, err := bleve.NewIndex(env.StoragePath, "")
	if err != nil {
		return cloudsearch.Config{}, err
	}
	results := bleve.NewBleveResultStorage(index)

	authService := auth.NewAuthenticator(
		authgateway.NewAuthGatewayClient(
			auth.DefaultGatewayUrl,
			env.HttpPort,
			&http.Client{
				Timeout: time.Second * 10,
			},
		),
	)

	registry := cloudsearch.NewRegistry()
	registry.RegisterAccountType(cloudsearch.Dropbox,
		search.WithCaching(search.Builder("dropbox", dropbox.NewSearch), enableCaching, results),
		auth.Builder(dropbox.NewAuthenticator()),
	)
	registry.RegisterAccountType(
		cloudsearch.Google,
		search.WithCaching(google.SearchBuilder, enableCaching, results),
		google.AuthBuilder(authService, accounts, auth.OauthRedirectUrlFor(env, cloudsearch.Google)),
	)
	registry.RegisterContentTypes(
		cloudsearch.Document,
		cloudsearch.Email,
		cloudsearch.File,
		cloudsearch.Folder,
		cloudsearch.Image,
		cloudsearch.Video,
	)

	multiSearch := cloudsearch.NewMultiSearch(
		env,
		accounts,
		results,
		registry,
		func(q cloudsearch.Query) []cloudsearch.ResultFilter {
			return []cloudsearch.ResultFilter{
				cloudsearch.SetId, // no better place to set this ugh
				cloudsearch.FilterNotInRange,
				cloudsearch.Dedup(q),
				cloudsearch.FilterContent,
			}
		},
	)

	// TODO move this side effect somewhere else
	go multiSearch.WatchTokens()

	return cloudsearch.Config{
		Env:             env,
		AccountsStorage: accounts,
		SearchEngine:    &multiSearch,
		Registry:        registry,
		ResultsStorage:  results,
		AuthService:     authService,
	}, nil
}
