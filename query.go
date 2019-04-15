package cloudsearch

import (
	"github.com/araddon/dateparse"
	"regexp"
	"strings"
	"time"
)

type SearchMode string

// search kinds: cache only, live only or everything
const (
	All   SearchMode = "all"
	Live  SearchMode = "live"
	Cache SearchMode = "cache"
)

var SupportedModes = []SearchMode{All, Live, Cache}

var SupportedModesStr []string

func init() {
	for _, s := range SupportedModes {
		SupportedModesStr = append(SupportedModesStr, string(s))
	}
}

type Query struct {
	RawText      string
	SearchMode   SearchMode
	Text         string // query without tokens
	Before       *time.Time
	After        *time.Time
	AccountTypes []AccountType
	ContentTypes []ContentType
	MaxResults   int
	SearchId     string
	// TODO search mode
	// TODO involving
	// TODO status (favorited)
	// TODO order
}

var beforeQuery = regexp.MustCompile(`\b(before):([-\/0-9]+)`)
var afterQuery = regexp.MustCompile(`\b(after):([-\/0-9]+)`)
var serviceQuery = regexp.MustCompile(`\b(service):([\w]+)`)
var serviceQuery2 = regexp.MustCompile(`\b@\[(service):([\w]+)\]`)
var modeQuery = regexp.MustCompile(`\b(mode):([\w]+)`)
var typeQuery = regexp.MustCompile(`\b(type):([\w]+)`)
var typeQuery2 = regexp.MustCompile(`\b@\[(type):([\w]+)\]`)

func ParseQuery(q string, searchId string, r *Registry) Query {
	stripped := q
	s, stripped := parseEnumItems(serviceQuery, r.SupportedAccountTypesStr(), stripped)
	s2, stripped := parseEnumItems(serviceQuery2, r.SupportedAccountTypesStr(), stripped)
	c, stripped := parseEnumItems(typeQuery, r.supportedContentTypesStr(), stripped)
	c2, stripped := parseEnumItems(typeQuery2, r.supportedContentTypesStr(), stripped)
	m, stripped := parseEnumItems(modeQuery, SupportedModesStr, stripped)
	if len(m) == 0 {
		m = []string{string(All)}
	}
	b, stripped := parseTime(beforeQuery, stripped)
	a, stripped := parseTime(afterQuery, stripped)
	stripped = strings.TrimSpace(stripped)

	return Query{
		RawText:      q,
		Text:         stripped,
		AccountTypes: accountTypes(concat(s, s2)),
		ContentTypes: contentTypes(concat(c, c2)),
		Before:       b,
		After:        a,
		MaxResults:   100,
		SearchMode:   SearchMode(m[0]),
		SearchId:     searchId,
	}
}

func CanHandle(query Query, accountType AccountType, contentTypes []ContentType) bool {
	return (len(query.AccountTypes) == 0 || accountTypeIncluded(query.AccountTypes, accountType)) &&
		(len(query.ContentTypes) == 0 || ContainsAnyType(query.ContentTypes, contentTypes))
}

func QueryFormattedTime(t time.Time) string {
	return t.Format("2006-01-02")
}

func concat(a []string, b []string) []string {
	return append(a, b...)
}

func accountTypes(c []string) []AccountType {
	var res []AccountType
	for _, s := range c {
		res = append(res, AccountType(s))
	}
	return res
}

func contentTypes(c []string) []ContentType {
	var res []ContentType
	for _, s := range c {
		res = append(res, ContentType(s))
	}
	return res
}

func parseEnumItems(regex *regexp.Regexp, supported []string, q string) ([]string, string) {
	res := []string{}

	t := regex.FindAllStringSubmatch(q, -1)
	for _, m := range t {
		w := strings.Title(m[2])
		if StringsContain(supported, w) {
			res = append(res, w)
		}
	}

	return res, regex.ReplaceAllString(q, "")
}

func parseTime(regex *regexp.Regexp, q string) (*time.Time, string) {
	t := regex.FindAllStringSubmatch(q, 1)
	if len(t) > 0 {
		t, err := dateparse.ParseAny(t[0][2])
		if err == nil {
			return &t, regex.ReplaceAllString(q, "")
		}
	}
	return nil, q
}

func accountTypeIncluded(list []AccountType, a AccountType) bool {
	for _, r := range list {
		if r == a {
			return true
		}
	}
	return false
}
