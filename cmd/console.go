package main

import (
	"flag"
	"fmt"
	"github.com/herval/cloudsearch"
	"github.com/herval/cloudsearch/action"
	"github.com/herval/cloudsearch/config"
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

	c, err := config.NewConfig(env, true)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	mode := flag.Arg(0)
	switch mode {
	case "accounts":
		op := flag.Arg(1)
		acc := flag.Arg(2)
		action.ListOrRemove(c.AccountsStorage, op, acc, *format)
	case "login":
		accType := flag.Arg(1)
		action.ConfigureNewAccount(
			accType,
			c.Env,
			c.AccountsStorage,
			c.Registry,
			c.AuthService,
		)
	case "search":
		searchString := strings.Join(flag.Args()[1:], " ")
		action.SearchAll(searchString, c.SearchEngine, c.Registry)
	default:
		if len(flag.Args()) == 0 {
			err := action.InteractiveMode(c.SearchEngine)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
		} else {
			flag.Usage()
		}
	}
}
