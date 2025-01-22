package config

import (
	"os"

	"github.com/sirupsen/logrus"
)

var Logger *logrus.Logger

// InitLogger initializes the Logrus logger
func InitLogger() {
	Logger = logrus.New()
	Logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	Logger.SetOutput(os.Stdout)
	Logger.SetLevel(logrus.InfoLevel)
}
