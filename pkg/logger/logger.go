package logger

import (
	"io"
	"log"
	"os"
)

type logger struct {
	warningLogger *log.Logger
	infoLogger    *log.Logger
	errorLogger   *log.Logger
	debugLogger   *log.Logger
}

var Log logger

func New(writer io.Writer) {
	Log = logger{}
	Log.infoLogger = log.New(writer, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	Log.warningLogger = log.New(writer, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	Log.errorLogger = log.New(writer, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	Log.debugLogger = log.New(writer, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func (l *logger) Info(msg ...interface{}) {
	l.infoLogger.Println(msg...)
}

func (l *logger) Warn(msg ...interface{}) {
	l.warningLogger.Println(msg...)
}

func (l *logger) Error(msg ...interface{}) {
	l.errorLogger.Println(msg...)
}

func (l *logger) Fatal(msg ...interface{}) {
	l.errorLogger.Fatalln(msg...)
}

func (l *logger) Debug(msg ...interface{}) {
	if os.Getenv("LOG_LEVEL") == "DEBUG" {
		l.debugLogger.Println(msg...)
	}
}
