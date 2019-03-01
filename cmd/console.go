package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/herval/cloudsearch"
	"github.com/herval/cloudsearch/gocui"
	"github.com/sirupsen/logrus"
	"os"
	"strings"
)

func main() {
	storagePath := flag.String("storagePath", "./", "storage path") // TODO . folder?
	oauthPort := flag.String("oauthPort", ":65432", "HTTP Port for Oauth2 callbacks")
	format := flag.String("format", "plain", "Output format for results (plain, json)")
	debug := flag.Bool("debug", false, "Debug logging")
	log := flag.Bool("log", false, "Output logging to a file")

	flag.Parse()

	if *debug {
		cloudsearch.LogLevel = logrus.DebugLevel
	} else {
		cloudsearch.LogLevel = logrus.InfoLevel
	}
	logrus.SetLevel(cloudsearch.LogLevel)

	if *log {
		f, err := os.OpenFile("cloudsearch.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
		if err != nil {
			panic("Could not setup log file")
		}
		logrus.SetOutput(f)
		logrus.Info("Starting application...")
	}

	config := NewConfig(*storagePath, *oauthPort, true)

	mode := flag.Arg(0)
	switch mode {
	case "accounts":
		op := flag.Arg(1)
		acc := flag.Arg(2)
		listOrRemove(config.AccountsStorage, op, acc, *format)
	case "login":
		accType := flag.Arg(1)
		configure(accType, config.Env, config.AccountsStorage, config.AuthBuilder, config.AuthService)
	case "search":
		searchString := strings.Join(flag.Args()[1:], " ")
		searchAll(searchString, config.SearchEngine, config.Env)
	default:
		if len(flag.Args()) == 0 {
			err := interactiveMode(config.SearchEngine)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
		} else {
			flag.Usage()
		}
	}
}

func interactiveMode(engine *cloudsearch.SearchEngine) error {
	return gocui.StartSearchApp(engine)
}

func startOauthServer(
	env cloudsearch.Env,
	accounts cloudsearch.AccountsStorage,
	auth cloudsearch.AuthBuilder,
	authService cloudsearch.OAuth2Authenticator,
) (<-chan error) {
	done := make(chan error)

	a := &cloudsearch.Api{
		env,
		accounts,
		auth,
		authService,
	}
	go func() {
		if err := a.Start(env.HttpPort, done); err != nil {
			logrus.Fatal("Could not start server", err)
			done <- err
		}
	}()

	return done
}

func listOrRemove(storage cloudsearch.AccountsStorage, op string, accountId string, format string) {
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

		switch format {
		case "plain":
			fmt.Println("Configured accounts:")
			for _, a := range accts {
				fmt.Println(fmt.Sprintf("%s - %s (%s)", a.ID, a.Description, a.AccountType))
			}
		case "json":
			//res := []string{}
			for _, a := range accts {
				// TODO format
				a.JsonFields()
				fmt.Println(fmt.Sprintf("%s - %s (%s)", a.ID, a.Description, a.AccountType))
			}
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
			logrus.Info("RESULT ->", q.Title)
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

	err := <-done
	if err != nil {
		fmt.Println("\nAuthentication failed: ", err.Error())
		os.Exit(1)
	}

	fmt.Println("\nAuthentication done!")
}
