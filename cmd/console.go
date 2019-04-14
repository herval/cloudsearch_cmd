package main

import (
	"flag"
	"fmt"
	"github.com/herval/cloudsearch"
	"github.com/herval/cloudsearch/action"
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

	err := cloudsearch.ConfigureLogging(*debug, *log)
	if err != nil {
		panic(err)
	}

	env := cloudsearch.Env{
		ServerBase:  "http://localhost",
		StoragePath: *storagePath,
		HttpPort:    *oauthPort,
	}

	config := NewConfig(env, true)

	mode := flag.Arg(0)
	switch mode {
	case "accounts":
		op := flag.Arg(1)
		acc := flag.Arg(2)
		action.ListOrRemove(config.AccountsStorage, op, acc, *format)
	case "login":
		accType := flag.Arg(1)
		action.ConfigureNewAccount(accType, config.Env, config.AccountsStorage, config.AuthBuilder, config.AuthService)
	case "search":
		searchString := strings.Join(flag.Args()[1:], " ")
		action.SearchAll(searchString, config.SearchEngine, config.Env)
	default:
		if len(flag.Args()) == 0 {
			err := action.InteractiveMode(config.SearchEngine)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
		} else {
			flag.Usage()
		}
	}
}
