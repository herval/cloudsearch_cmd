package dropbox

import "time"

type Content struct {
	Id          string
	Name        string
	Revision    string
	Hash        string
	ContentType string
	Path        string
	Body        string
	IsDir       bool
	Modified    time.Time
	Size        int64
}
