package dropbox

import (
	"github.com/herval/cloudsearch"
	"context"
	"fmt"
	"github.com/herval/dropbox-sdk-go-unofficial/dropbox"
	"github.com/herval/dropbox-sdk-go-unofficial/dropbox/files"
	"github.com/sirupsen/logrus"
	"path"
	"time"
)

func NewSearch(account cloudsearch.AccountData) cloudsearch.SearchFunc {
	c := dropbox.Config{
		Token:    account.Token,
		LogLevel: dropbox.LogOff,
	}
	db := files.New(c)
	db.HttpClient().Timeout = time.Second * 2

	s := searchable{
		db:      db,
		account: account,
	}

	return s.SearchSnippets
}

func (s *searchable) SearchSnippets(query cloudsearch.Query, ctx context.Context) <-chan cloudsearch.Result {
	out := make(chan cloudsearch.Result)

	if !cloudsearch.CanHandle(query, s.account.AccountType, cloudsearch.FileTypes) {
		close(out)
		return out
	}

	if ctx.Err() != nil {
		close(out)
		return out
	}

	go func() {
		defer close(out)
		res, err := s.search(query.Text)
		if err != nil {
			logrus.Trace("Error searching:", err)
			return
		} else {
			for _, r := range s.toResults(res) {
				out <- r
			}
		}
	}()

	return out
}

func (s *searchable) toResults(contents []Content) []cloudsearch.Result {
	res := make([]cloudsearch.Result, len(contents))
	for i, e := range contents {
		res[i] = cloudsearch.FileOrFolderResult(
			e.Id,
			e.Path,
			e.Path,
			path.Ext(e.Path),
			"",
			e.Modified,
			fmt.Sprintf(
				"https://www.dropbox.com/home%s?preview=%s",
				e.Path,
				e.Name,
			),
			e.Size,
			"", // TODO ?
			s.account,
			"", // TODO ?
			true,
			[]string{},
			e.IsDir,
		)
	}
	return res
}

func (s *searchable) search(query string) ([]Content, error) {
	logrus.Trace("Searching:", query)

	res, err := s.db.Search(&files.SearchArg{
		Path:       "",
		Query:      query,
		MaxResults: 100, // TODO configure?
		Mode: &files.SearchMode{
			Tagged: dropbox.Tagged{
				Tag: files.SearchModeFilenameAndContent,
			},
		},
	})

	if err != nil {
		return nil, err
	}

	results := make([]Content, len(res.Matches))
	for i, r := range res.Matches {
		c := convert(r)
		if c != nil {
			results[i] = *c
		}
	}

	return results, nil
}

func convert(e *files.SearchMatch) *Content {
	switch t := e.Metadata.(type) {
	case *files.FolderMetadata:
		return &Content{
			Id:    t.Id,
			Path:  t.PathLower,
			IsDir: true,
		}
	case *files.FileMetadata:
		return &Content{
			Id:       t.Id,
			Path:     t.PathLower,
			Hash:     t.ContentHash,
			Revision: t.Rev,
			IsDir:    false,
			Modified: cloudsearch.Latest(t.ServerModified, t.ClientModified),
			Size:     int64(t.Size),
			Name:     t.Name,
		}
	default:
		return nil
	}
}

type searchable struct {
	db      files.Client
	account cloudsearch.AccountData
}
