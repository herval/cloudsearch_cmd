package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/herval/cloudsearch"
	"github.com/sirupsen/logrus"
	"os"
	"strings"
)

func main() {
	storagePath := flag.String("storagePath", "./", "storage path") // TODO . folder?
	oauthPort := flag.String("oauthPort", ":65432", "HTTP Port for Oauth2 callbacks")
	debug := flag.Bool("debug", false, "Debug logging")
	flag.Parse()

	if *debug {
		cloudsearch.LogLevel = logrus.DebugLevel
	} else {
		cloudsearch.LogLevel = logrus.InfoLevel
	}
	logrus.SetLevel(cloudsearch.LogLevel)

	mode := flag.Arg(0)
	op := flag.Arg(1)

	config := NewConfig(*storagePath, *oauthPort, true)

	switch mode {
	case "accounts":
		listOrRemove(config.AccountsStorage, op, flag.Arg(2))
	case "login":
		configure(op, config.Env, config.AccountsStorage, config.AuthBuilder, config.AuthService)
	case "search":
		searchAll(strings.Join(flag.Args()[1:], " "), config.SearchEngine, config.Env)
	default:
		flag.Usage()
	}
}

func startOauthServer(
	env cloudsearch.Env,
	accounts cloudsearch.AccountsStorage,
	auth cloudsearch.AuthBuilder,
	authService cloudsearch.OAuth2Authenticator,
) (<-chan bool) {
	done := make(chan bool)

	a := &api{
		env,
		accounts,
		auth,
		authService,
	}
	go func() {
		if err := a.Start(env.HttpPort, done); err != nil {
			logrus.Fatal("Could not start server", err)
			done <- true
		}
	}()

	return done
}

func listOrRemove(storage cloudsearch.AccountsStorage, op string, accountId string) {
	switch op {
	case "list":
		accts, err := storage.All()
		if err != nil {
			panic(err)
		}
		if len(accts) == 0 {
			fmt.Println("No accounts configured - use 'cloudsearch login <provider>' to register one!")
			return
		}
		
		fmt.Println("Configured accounts:")
		for _, a := range accts {
			fmt.Println(fmt.Sprintf("%s - %s (%s)", a.ID, a.Description, a.AccountType))
		}
	case "remove":
		err := storage.Delete(accountId)
		if err != nil {
			fmt.Println("Could not remove account: ", err)
			os.Exit(1)
		}

		fmt.Println("Account removed!")
		os.Exit(0)
	default:
		fmt.Println("Please provide a valid operation. (list | remove).\nExample usage:\n> cloudsearch accounts list\n> cloudsearch accounts remove 123456")
		os.Exit(1)
	}
}

func searchAll(cmd string, search *cloudsearch.SearchEngine, config cloudsearch.Env) {
	query := cloudsearch.ParseQuery(cmd, cloudsearch.NewId())
	res := search.Search(query, context.Background())

	for {
		select {
		case q, ok := <-res:
			if !ok {
				logrus.Debug("All done!")
				os.Exit(0)
			}
			logrus.Info("RESULT ->", q)
		}
	}
}

func configure(
	accType string,
	env cloudsearch.Env,
	storage cloudsearch.AccountsStorage,
	authBuilder cloudsearch.AuthBuilder,
	authenticator cloudsearch.OAuth2Authenticator,
) {
	acc := cloudsearch.AccountType(accType)
	if !cloudsearch.AccountTypeIncluded(cloudsearch.SupportedAccountTypes, acc) {
		fmt.Println("Please provide a valid account type (Dropbox | Google).\nExample usage:\n> cloudsearch login Google")
		os.Exit(1)
	}

	done := startOauthServer(env, storage, authBuilder, authenticator)

	url := fmt.Sprintf("%s%s/oauth/start/%s", env.ServerBase, env.HttpPort, accType)
	fmt.Println("To start the login flow for your account, go to this link:\n\n    " + url + "\n")

	<-done

	fmt.Println("\nAuthentication done!")
}
