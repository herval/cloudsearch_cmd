package cloudsearch

import (
	"github.com/sirupsen/logrus"
	"os"
)

var LogLevel = logrus.DebugLevel

func ConfigureLogging(debug bool, saveToFile bool) error {
	if debug {
		LogLevel = logrus.DebugLevel
	} else {
		LogLevel = logrus.InfoLevel
	}
	logrus.SetLevel(LogLevel)

	if saveToFile {
		f, err := os.OpenFile("cloudsearch.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
		if err != nil {
			return err
		}
		logrus.SetOutput(f)
		logrus.Info("Starting application...")
	}

	return nil
}
