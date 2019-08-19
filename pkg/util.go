package cloudsearch

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/google/uuid"
	"math"
	"strconv"
	"strings"
	"time"
)

func NewId() string {
	return Md5(uuid.New().String())
}

func FileAt(path string, file string) string {
	if strings.HasSuffix(path, "/") {
		return path + file
	} else {
		return path + "/" + file
	}
}

func TimeFromMillis(millis int64) time.Time {
	return time.Unix(0, millis*int64(time.Millisecond))
}

func TimeFromFloatMillis(ts string) time.Time {
	tsFloat, _ := strconv.ParseFloat(ts, 64) // TODO handle errors?
	sec, dec := math.Modf(tsFloat)
	return time.Unix(int64(sec), int64(dec*(1e9)))
}

func Latest(t1 time.Time, t2 time.Time) time.Time {
	if t1.After(t2) {
		return t1
	}
	return t2
}

func Either(str1 string, str2 string) string {
	if str1 == "" {
		return str2
	}
	return str1
}

func ParseOrNil(t string, format string) time.Time {
	r, _ := time.Parse(format, t)
	return r
}

func StringsContain(strs []string, str string) bool {
	for _, s := range strs {
		if s == str {
			return true
		}
	}
	return false
}

func Md5(str string) string {
	hasher := md5.New()
	hasher.Write([]byte(str))
	return hex.EncodeToString(hasher.Sum(nil))
}
