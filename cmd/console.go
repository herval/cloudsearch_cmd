package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/herval/cloudsearch"
	"github.com/sirupsen/logrus"
	"os"
)

func main() {
	storagePath := flag.String("storagePath", "./", "storage path") // TODO . folder?
	oauthPort := flag.String("oauthPort", ":65432", "HTTP Port for Oauth2 callbacks")
	debug := flag.Bool("debug", false, "Debug logging")

	if *debug {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	//account := flag.String("account", "", "configure a new account (Dropbox | Google)")
	//op := flag.String("op", "", "additional params for specific operations")
	flag.Parse()

	mode := flag.Arg(0)
	op := flag.Arg(1)
	//mode := flag.String("mode", "", "run mode (setup | search | accounts | server)")

	config := NewConfig(*storagePath, *oauthPort, true)

	switch mode {
	case "accounts":
		listOrRemove(config.AccountsStorage, op)
	case "login":
		configure(op, config.Env, config.AccountsStorage, config.AuthBuilder, config.AuthService)
	case "search":
		searchAll(op, config.SearchEngine, config.Env)
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

func listOrRemove(storage cloudsearch.AccountsStorage, op string) {
	switch op {
	case "list":
		fmt.Println("Configured accounts:")
		accts, err := storage.All()
		if err != nil {
			panic(err)
		}
		for _, a := range accts {
			fmt.Println(fmt.Sprintf("%s - %s", a.ID, a.Name))
		}
	case "remove":
		// TODO
		os.Exit(1)
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
				logrus.Info("All done!")
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
