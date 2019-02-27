package google

import (
	"github.com/herval/cloudsearch"
	"net/http"
)

func SearchablesFor(httpClient *http.Client, account cloudsearch.AccountData) ([]cloudsearch.SearchFunc, []string, error) {
	drive, err := NewGoogleDrive(account, httpClient)
	if err != nil {
		return nil, nil, err
	}

	gmail, err := NewGmail(account, httpClient)
	if err != nil {
		return nil, nil, err
	}

	return []cloudsearch.SearchFunc{
		drive.SearchSnippets,
		gmail.SearchSnippets,
	}, []string{
		"drive",
		"gmail",
	}, nil
}
