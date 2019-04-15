package main

import (
	"github.com/herval/authgateway/client"
	"github.com/herval/cloudsearch"
	"github.com/herval/cloudsearch/auth"
	"github.com/herval/cloudsearch/search"
	"github.com/herval/cloudsearch/search/dropbox"
	"github.com/herval/cloudsearch/search/google"
	"github.com/herval/cloudsearch/storage/bleve"
	"github.com/herval/cloudsearch/storage/storm"
	"net/http"
	"time"
)

func NewConfig(env cloudsearch.Env, enableCaching bool) cloudsearch.Config {
	accounts, err := storm.NewAccountsStorage(env.StoragePath)
	if err != nil {
		panic(err)
	}

	index, err := bleve.NewIndex(env.StoragePath, "")
	if err != nil {
		panic(err)
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
		search.Builder("dropbox", dropbox.NewSearch),
		auth.Builder(dropbox.NewAuthenticator()),
	)
	registry.RegisterAccountType(
		cloudsearch.Google,
		google.SearchBuilder,
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

	searchBuilder := registry.SearchBuilder
	if enableCaching {
		searchBuilder = search.NewCachedSearchableBuilder(results, searchBuilder, enableCaching)
	}

	multiSearch := cloudsearch.NewMultiSearch(
		env,
		accounts,
		results,
		registry,
		func(q cloudsearch.Query) []cloudsearch.ResultFilter {
			return []cloudsearch.ResultFilter{
				cloudsearch.FilterNotInRange,
				cloudsearch.Dedup(q),
				cloudsearch.FilterContent,
			}
		},
	)

	go multiSearch.WatchTokens()

	return cloudsearch.Config{
		Env:             env,
		AccountsStorage: accounts,
		SearchEngine:    &multiSearch,
		Registry:        registry,
		ResultsStorage:  results,
		AuthService:     authService,
	}
}
