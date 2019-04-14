package google

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"net/http"
	"strings"
	"time"

	"github.com/herval/cloudsearch"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/drive/v3"
)

type GoogleDrive struct {
	driveApi     *drive.Service
	googleClient *http.Client
	account      cloudsearch.AccountData
}

func NewGoogleDrive(
	account cloudsearch.AccountData,
	client *http.Client,
) (*GoogleDrive, error) {

	drv, err := drive.New(client)

	return &GoogleDrive{
		account:  account,
		driveApi: drv,
	}, err
}

func (a *GoogleDrive) Search(ctx context.Context, q string, pageToken string) (*drive.FileList, string, error) {
	//logrus.Debug("gdrive: ", q, " - token: ", pageToken)
	r, err := a.driveApi.Files.
		List().
		Q(q).
		PageSize(100).
		PageToken(pageToken).
		Context(ctx).
		Fields("files(id,name,size,createdTime,modifiedTime,thumbnailLink,webViewLink,fileExtension,mimeType,iconLink)").
		Do()
	if err != nil {
		return nil, "", errors.Wrap(err, "searching gdrive")
	}
	return r, r.NextPageToken, nil
}

func (a *GoogleDrive) SearchSnippets(query cloudsearch.Query, ctx context.Context) <-chan cloudsearch.Result {
	out := make(chan cloudsearch.Result)

	if !cloudsearch.CanHandle(query, a.account.AccountType, cloudsearch.FileTypes) {
		close(out)
		return out
	}

	q := a.buildQuery(query)
	if ctx.Err() != nil {
		close(out)
		return out
	}

	go func() {
		defer close(out)
		r, _, err := a.Search(ctx, q, "")
		if err != nil {
			//logrus.Error("Couldn't list:", err)
			return
		} else {
			for _, f := range r.Files {
				if ctx.Err() != nil {
					return
				}

				out <- a.ToResult(f)
			}
		}
	}()

	return out
}

func (a *GoogleDrive) buildQuery(query cloudsearch.Query) string {
	q := fmt.Sprintf("fullText contains '%s'", query.Text)

	// modifiedTime modifiedTime > '2012-06-04T12:00:00'
	// owners writers readers in
	if query.After != nil {
		q += fmt.Sprintf(" and modifiedTime > '%s'", a.FormattedTime(*query.After))
	}
	if query.Before != nil {
		q += fmt.Sprintf(" and modifiedTime < '%s'", a.FormattedTime(*query.Before))
	}

	if query.ContentTypes != nil {
		types := []string{}

		for _, t := range query.ContentTypes {
			if t == cloudsearch.Image {
				types = append(types, "(mimeType contains 'image') or (mimeType contains 'drawing')")
			} else if t == cloudsearch.Video {
				types = append(types, "(mimeType contains 'video')")
			} else if t == cloudsearch.Folder {
				types = append(types, "(mimeType contains 'folder')")
			} else if t == cloudsearch.Document {
				types = append(types, "(mimeType contains 'document')")
			} else if t == cloudsearch.File {
				types = append(types, "(mimeType contains 'file')")
			}
		}

		if len(types) > 0 {
			q += " and (" + strings.Join(types, " or ") + ")"
		}
	}

	logrus.Debug("Searching GDrive: ", q)

	return q
}

func (a *GoogleDrive) FormattedTime(t time.Time) string {
	//2012-06-04T12:00:00
	return t.Format(time.RFC3339)
}

func (a *GoogleDrive) ToResult(f *drive.File) cloudsearch.Result {
	return cloudsearch.FileOrFolderResult(
		f.Id,
		cloudsearch.Either(f.OriginalFilename, f.Name),
		f.Name,
		f.FileExtension,
		f.MimeType,
		cloudsearch.ParseOrNil(
			cloudsearch.Either(
				f.ModifiedTime,
				f.CreatedTime,
			),
			time.RFC3339),
		f.WebViewLink,
		f.Size,
		"",
		a.account,
		cloudsearch.Either(
			f.ThumbnailLink,
			f.IconLink,
		),
		true, // TODO parse from query
		[]string{},
		false,
	)
}
