package handlers

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger

func init() {

	initLogger()

}

func initLogger() {
	writerStd := os.Stdout
	writerFile, err := os.OpenFile("logs.txt", os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		logrus.Fatalf("create file log.txt failed: %v", err)
	}

	logger = logrus.New()
	// logger.SetLevel(logrus.InfoLevel)
	logger.SetLevel(logrus.DebugLevel)
	logger.SetFormatter(&logrus.TextFormatter{})
	logger.SetOutput(io.MultiWriter(writerStd, writerFile))
}

func GetLogger() *logrus.Logger {
	return logger
}
