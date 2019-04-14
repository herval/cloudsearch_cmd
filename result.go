package cloudsearch

import (
	"fmt"
	"time"
)

type ResultStatus int

// order matters! (zero value == ResultFound)
const (
	ResultFound ResultStatus = iota
	ResultNotFound
	ResultError
)

type Result struct {
	Id            string
	AccountId     string
	AccountType   AccountType
	Title         string
	Permalink     string
	Thumbnail     string
	Timestamp     time.Time
	ContentType   ContentType
	Details       map[string]interface{}
	OriginalId    string
	Body          string
	CachedAt      time.Time
	Labels        []string
	InvolvesMe    bool // little hack to differentiate involves:anyone from involves:me
	Status        ResultStatus
	Unread        bool
	CacheHitScore float64 `json:"-"` // a transient hit score of the result, based on relevance on cache _only_
	// TODO involved
}

func FileOrFolderResult(
	originalId string,
	path string,
	title string,
	extension string,
	mimeType string,
	timestamp time.Time,
	permalink string,
	sizeBytes int64,
	body string,
	account AccountData,
	thumbnail string,
	involvesMe bool,
	labels []string,
) Result {
	klass := kindFor(mimeType, extension, path)

	switch klass {
	case Image:
		return ImageResult(originalId, path, title, timestamp, permalink, body, account, sizeBytes, thumbnail, involvesMe, labels)
	case Video:
		return VideoResult(originalId, path, title, timestamp, permalink, body, account, sizeBytes, thumbnail, involvesMe, labels)
	case Document:
		return DocumentResult(originalId, path, title, timestamp, permalink, body, account, sizeBytes, thumbnail, involvesMe, labels)
	case Folder:
		return FolderResult(originalId, path, title, timestamp, permalink, body, account, sizeBytes, thumbnail, involvesMe, labels)
	default:
		return BasicFileResult(originalId, path, title, timestamp, permalink, body, account, sizeBytes, thumbnail, involvesMe, labels)
	}
}

func FolderResult(originalId string, path string, title string, timestamp time.Time, permalink string, body string, accountData AccountData, sizeBytes int64, thumbnail string, involvesMe bool, labels []string) Result {
	return Result{
		AccountId:   accountData.ID,
		AccountType: accountData.AccountType,
		Title:       title,
		Permalink:   permalink,
		Thumbnail:   thumbnail,
		Timestamp:   timestamp,
		ContentType: Folder,
		Details: map[string]interface{}{
			"path": path,
		},
		OriginalId: originalId,
		Labels:     labels,
		InvolvesMe: involvesMe,
	}
}

func VideoResult(originalId string, path string, title string, timestamp time.Time, permalink string, body string, accountData AccountData, sizeBytes int64, thumbnail string, involvesMe bool, labels []string) Result {
	return Result{
		AccountId:   accountData.ID,
		AccountType: accountData.AccountType,
		Title:       title,
		Permalink:   permalink,
		Thumbnail:   thumbnail,
		Timestamp:   timestamp,
		ContentType: Video,
		Details: map[string]interface{}{
			"sizeBytes": sizeBytes,
			"path":      path,
		},
		OriginalId: originalId,
		Labels:     labels,
		InvolvesMe: involvesMe,
	}
}

func BasicFileResult(originalId string, path string, title string, timestamp time.Time, permalink string, body string, accountData AccountData, sizeBytes int64, thumbnail string, involvesMe bool, labels []string) Result {
	return Result{
		AccountId:   accountData.ID,
		AccountType: accountData.AccountType,
		Title:       title,
		Permalink:   permalink,
		Thumbnail:   thumbnail,
		Timestamp:   timestamp,
		ContentType: File,
		Details: map[string]interface{}{
			"sizeBytes": sizeBytes,
			"path":      path,
		},
		OriginalId: originalId,
		Labels:     labels,
		InvolvesMe: involvesMe,
	}
}

func DocumentResult(originalId string, path string, title string, timestamp time.Time, permalink string, body string, accountData AccountData, sizeBytes int64, thumbnail string, involvesMe bool, labels []string) Result {
	return Result{
		AccountId:   accountData.ID,
		AccountType: accountData.AccountType,
		Title:       title,
		Permalink:   permalink,
		Thumbnail:   thumbnail,
		Timestamp:   timestamp,
		ContentType: Document,
		Details: map[string]interface{}{
			"sizeBytes": sizeBytes,
			"path":      path,
		},
		OriginalId: originalId,
		Body:       "",
		Labels:     labels,
		InvolvesMe: involvesMe,
	}
}

func ImageResult(originalId string, path string, title string, timestamp time.Time, permalink string, body string, accountData AccountData, sizeBytes int64, thumbnail string, involvesMe bool, labels []string) Result {
	return Result{
		AccountId:   accountData.ID,
		AccountType: accountData.AccountType,
		Title:       title,
		Permalink:   permalink,
		Thumbnail:   thumbnail,
		Timestamp:   timestamp,
		ContentType: Image,
		Details: map[string]interface{}{
			"sizeBytes": sizeBytes,
			"path":      path,
		},
		OriginalId: originalId,
		Body:       body,
		Labels:     labels,
		InvolvesMe: involvesMe,
	}
}

// TODO content relevance depending on query?
func (r *Result) Relevance(q Query) int64 {
	switch r.ContentType {
	case Event:
		ts := r.Timestamp
		res := standardRelevance(r)
		if ts.After(time.Now()) {
			if ts.Before(time.Now().Add(time.Hour * 24 * 2)) {
				res *= 10
			} else if ts.Before(time.Now().Add(time.Hour * 24 * 5)) {
				res *= 2
			} else {
				res /= 5
			}
		} else {
			res /= 100
		}
		return res

	case Contact:
		return standardRelevance(r) * 2 // contacts are always important...?
	case Folder:
		return standardRelevance(r) / 100 // folders aren't relevant
	default:
		return standardRelevance(r)
	}
	//            ContentType.Task -> // TODO non-closed scores higher, with deadline scores higher x2
}

func (r *Result) SetId() {
	r.Id = Md5(fmt.Sprintf("%s_%s_%s", r.AccountId, r.OriginalId, r.ContentType))
}

func standardRelevance(c *Result) int64 {
	// cached items have a lower relevance than live searched ones
	//logrus.Debug(c.Id, " - score: ", c.CacheHitScore, " - favorited: ", c.Favorited)
	// TODO if we apply a score to non-cached results, we'll normalize search noise (treat them as "if I searched in cache" scores)
	// TODO save a score bump on items that get clicked, so they rank higher next times for the same query

	ret := c.Timestamp.Unix() / 100000

	//if c.CacheHitScore > 0.0 {
	//	ret = int64(float64(ret) * c.CacheHitScore)
	//}

	return ret
}
