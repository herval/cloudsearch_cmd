package action

import (
	"fmt"
	"github.com/herval/cloudsearch"
	"github.com/herval/cloudsearch/auth"
	"github.com/sirupsen/logrus"
	"os"
)

func ConfigureNewAccount(
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

	done := StartOauthServer(env, storage, authBuilder, authenticator)

	url := fmt.Sprintf("%s%s/oauth/start/%s", env.ServerBase, env.HttpPort, accType)
	fmt.Println("To start the login flow for your account, go to this link:\n\n    " + url + "\n")

	err := <-done
	if err != nil {
		fmt.Println("\nAuthentication failed: ", err.Error())
		os.Exit(1)
	}

	fmt.Println("\nAuthentication done!")
}


func ListOrRemove(storage cloudsearch.AccountsStorage, op string, accountId string, format string) {
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

func StartOauthServer(
	env cloudsearch.Env,
	accounts cloudsearch.AccountsStorage,
	ab cloudsearch.AuthBuilder,
	authService cloudsearch.OAuth2Authenticator,
) (<-chan error) {
	done := make(chan error)

	a := &auth.Api{
		env,
		accounts,
		ab,
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


