package cloudsearch

import (
	"github.com/sirupsen/logrus"
	"time"
)

type Stopwatch struct {
	name  string
	start time.Time
}

func NewStopwatch(op string) Stopwatch {
	return Stopwatch{
		name:  op,
		start: time.Now(),
	}
}

func (t Stopwatch) Lap() time.Duration {
	done := time.Since(t.start)
	logrus.Debug(t.name, " finished in ", done)

	return done
}
