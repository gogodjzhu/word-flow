package log

import (
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"path"
)

var MainLogger *log.Logger

func init() {
	MainLogger = log.New()
	MainLogger.Out = createLogWriter("/var/log/wordflow", "main.log")
	MainLogger.SetLevel(log.DebugLevel)
	MainLogger.SetFormatter(&log.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
	})
	MainLogger.Logf(log.DebugLevel, "MainLogger init success")
}

func createLogWriter(logFilePath, logFileName string) io.Writer {
	// if env is dev, use stdout
	if os.Getenv("ENV") == "dev" {
		return createStdoutWriter()
	} else {
		return createOrOpenLogFileWriter(logFilePath, logFileName)
	}
}

func createStdoutWriter() io.Writer {
	return os.Stdout
}

func createOrOpenLogFileWriter(logFilePath, logFileName string) io.Writer {
	err := os.MkdirAll(logFilePath, 0755)
	if err != nil {
		panic("Failed to create log file" + err.Error())
	}
	fileName := path.Join(logFilePath, logFileName)
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		_, err = os.Create(fileName)
		if err != nil {
			panic("Failed to create log file" + err.Error())
		}
	}
	src, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		panic("Failed to open log file" + err.Error())
	}
	return src
}
