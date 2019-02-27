package main

import (
	"github.com/herval/authgateway/client"
	"github.com/herval/cloudsearch"
	"github.com/herval/cloudsearch/search/authenticator"
	"github.com/herval/cloudsearch/storage/bleve"
	"github.com/herval/cloudsearch/storage/storm"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type Config struct {
	Env              cloudsearch.Env
	AccountsStorage  cloudsearch.AccountsStorage
	SearchEngine     *cloudsearch.SearchEngine
	AuthBuilder      cloudsearch.AuthBuilder
	ResultsStorage   cloudsearch.ResultsStorage
	AuthService      cloudsearch.OAuth2Authenticator
}

func NewConfig(storagePath string, httpPort string, enableCaching bool) Config {
	env := cloudsearch.Env{
		ServerBase:  "http://localhost",
		StoragePath: storagePath,
		HttpPort:    httpPort,
	}

	// TODO setup logging
	logrus.SetLevel(logrus.DebugLevel)

	accounts, err := storm.NewAccountsStorage(storagePath)
	if err != nil {
		panic(err)
	}

	index, err := bleve.NewIndex(storagePath, "")
	if err != nil {
		panic(err)
	}
	results := bleve.NewBleveResultStorage(index)


	authService := authenticator.NewAuthenticator(
		authgateway.NewAuthGatewayClient(
			authenticator.DefaultGatewayUrl,
			env.HttpPort,
			&http.Client{
				Timeout: time.Second * 10,
			},
		),
	)

	auth := NewAuthBuilder(
		authService,
		accounts,
		OauthRedirectUrlFor(env, cloudsearch.Google),
	)

	searchBuilder := NewRemoteSearchablesBuilder(auth)
	if enableCaching {
		searchBuilder = NewCachedSearchableBuilder(results, searchBuilder, enableCaching)
	}

	multiSearch := cloudsearch.NewMultiSearch(
		env,
		accounts,
		results,
		searchBuilder,
		auth,
		func(q cloudsearch.Query) []cloudsearch.ResultFilter {
			return []cloudsearch.ResultFilter{
				cloudsearch.FilterNotInRange,
				cloudsearch.Dedup,
				cloudsearch.FilterContent,
			}
		},
	)

	go multiSearch.WatchTokens()

	return Config{
		env,
		accounts,
		&multiSearch,
		auth,
		results,
		authService,
	}
}
