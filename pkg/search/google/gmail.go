package google

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/pkg/errors"
	"google.golang.org/api/gmail/v1"
	"net/http"
	"strings"
	"time"

	"github.com/herval/cloudsearch/pkg"
	"github.com/sirupsen/logrus"
)

type Gmail struct {
	api          *gmail.Service
	googleClient *http.Client
	account      cloudsearch.AccountData
}

var GmailTimeFormat = "2006/01/02"

func NewGmail(
	account cloudsearch.AccountData,
	client *http.Client,
) (*Gmail, error) {

	drv, err := gmail.New(client)

	return &Gmail{
		account: account,
		api:     drv,
	}, err
}

func (a *Gmail) Search(ctx context.Context, q string, pageToken string, out chan<- cloudsearch.Result) (string, error) {
	//logrus.Debug("gmail: ", q, " - token: ", pageToken)
	r, err := a.api.Users.Messages.
		List(a.account.Email).
		Q(q).
		PageToken(pageToken).
		Context(ctx).
		Fields("messages(id)").
		Do()
	if err != nil {
		return "", errors.Wrap(err, "searching gmail")
	}

	for _, m := range r.Messages {
		if ctx.Err() != nil {
			//logrus.Debug("Cancelling gmail search...")
			return r.NextPageToken, ctx.Err()
		}

		m, err := a.api.Users.Messages.
			Get(a.account.Email, m.Id).
			Format("full").
			Do()
		if err != nil {
			return "", err
			//logrus.Error("Couldn't fetch message: ", err.Error())
		} else {
			out <- a.toResult(m)
		}
	}

	return r.NextPageToken, nil
}

func (a *Gmail) SearchSnippets(query cloudsearch.Query, ctx context.Context) <-chan cloudsearch.Result {
	out := make(chan cloudsearch.Result)

	if !cloudsearch.CanHandle(query, a.account.AccountType, []cloudsearch.ContentType{cloudsearch.Email}) {
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
		_, err := a.Search(ctx, q, "", out)
		if err != nil {
			//logrus.Error("Couldn't list:", err)
			return
		}
	}()

	return out
}

func (a *Gmail) buildQuery(query cloudsearch.Query) string {
	q := fmt.Sprintf("%s", query.Text)

	// modifiedTime modifiedTime > '2012-06-04T12:00:00'
	// owners writers readers in
	if query.After != nil {
		q += fmt.Sprintf(" and after > '%s'", a.FormattedTime(*query.After))
	}
	if query.Before != nil {
		q += fmt.Sprintf(" and before < '%s'", a.FormattedTime(*query.Before))
	}

	// TODO has:attachment

	logrus.Debug("Searching Gmail: ", q)

	return q
}

func (a *Gmail) FormattedTime(t time.Time) string {
	return t.Format(GmailTimeFormat)
}

func (a *Gmail) toResult(f *gmail.Message) cloudsearch.Result {
	//logrus.Debug(fmt.Sprintf("%+v", f.Header))

	subject := ""
	recipients := []string{}
	from := []string{}
	to := []string{}
	for _, h := range f.Payload.Headers {
		//logrus.Debug(h)
		if h.Name == "To" {
			to = append(to, h.Value)
		}

		if h.Name == "From" {
			from = append(from, h.Value)
		}

		if h.Name == "Subject" {
			subject = h.Value
		}
	}

	// TODO other mime types indicate files, which we can index separately
	body := ""
	for _, part := range f.Payload.Parts {
		//logrus.Debug(part.MimeType)
		if part.MimeType == "text/html" || part.MimeType == "text/plain" || part.MimeType == "text/xml" {
			data, err := base64.StdEncoding.DecodeString(part.Body.Data)
			if err == nil {
				body += string(data)
			}
		}
	}

	if strings.Trim(body, " ") == "" {
		body = f.Snippet
	}

	//logrus.Debug(body)

	recipients = append(recipients, to...)
	recipients = append(recipients, from...)

	labels := f.LabelIds
	//logrus.Debug(fmt.Sprintf("%s %s %s", subject, f.Snippet, recipients))

	unread := cloudsearch.StringsContain(labels, "UNREAD")

	involvesMe := cloudsearch.StringsContain(labels, "SENT") || cloudsearch.StringsContain(recipients, a.account.Email)

	return cloudsearch.Result{
		AccountId:   a.account.ID,
		AccountType: a.account.AccountType,
		Title:       subject,
		Permalink:   fmt.Sprintf("https://mail.google.com/mail/u/%s/#inbox/%s", a.account.Email, f.ThreadId),
		Thumbnail:   "",
		ContentType: cloudsearch.Email,
		OriginalId:  f.Id,
		Timestamp:   time.Unix(f.InternalDate/1000, 0),
		Body:        fmt.Sprintf("%s %s %s %s", subject, body, strings.Join(recipients, " "), strings.Join(labels, " ")), // TODO store the actual body here and get rid of the details one so we can highlight properly
		Unread:      unread,
		InvolvesMe:  involvesMe,
		Details: map[string]interface{}{
			"labels":  labels,
			"from":    strings.Join(from, ", "),
			"to":      strings.Join(to, ", "),
			"subject": subject,
			"body":    body,
		},
	}
}
