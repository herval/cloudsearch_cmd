package storm_test

import (
	"github.com/herval/cloudsearch/pkg/storage/storm"
	"testing"

	"github.com/herval/cloudsearch/pkg"
)

func TestStorage(t *testing.T) {
	env := cloudsearch.Env{
		StoragePath: "./../tmp",
	}
	storage, err := storm.NewAccountsStorage(env.StoragePath)
	if err != nil {
		t.Fail()
	}

	// cleanup
	accs, err := storage.All()
	if err != nil {
		t.Fatal(err)
	}
	for _, a := range accs {
		err = storage.Delete(a.ID)
		if err != nil {
			t.Fatal(err)
		}
	}

	// save a bunch of new accounts
	dropb := cloudsearch.AccountData{
		Name:        "123",
		ExternalId:  "123",
		Token:       "dropboxtoken",
		AccountType: cloudsearch.Dropbox,
	}
	googl := cloudsearch.AccountData{
		Name:        "456",
		ExternalId:  "456",
		Token:       "googletoken",
		AccountType: cloudsearch.Google,
	}
	googl2 := cloudsearch.AccountData{
		Name:        "678",
		ExternalId:  "678",
		Token:       "googletoken2",
		AccountType: cloudsearch.Google,
	}

	if err := storage.Save(&googl); err != nil {
		t.Fatal(err)
	}
	if err := storage.Save(&dropb); err != nil {
		t.Fatal(err)
	}
	if err := storage.Save(&googl2); err != nil {
		t.Fatal(err)
	}

	accts, err := storage.All()
	if err != nil {
		t.Fatal("Couldn't retrieve accounts:", err)
	}

	if len(accts) != 3 {
		t.Fatal("Expected accounts not found:", accts)
	}

	//fmt.Println(accts)
}
