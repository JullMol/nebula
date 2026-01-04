package logger

import (
	"log"
	"os"
)

var (
	Info  = log.New(os.Stdout, "\u001b[34m[INFO]\u001b[0m ", log.LstdFlags|log.Lshortfile)
	Error = log.New(os.Stderr, "\u001b[31m[ERROR]\u001b[0m ", log.LstdFlags|log.Lshortfile)
	Warn  = log.New(os.Stdout, "\u001b[33m[WARN]\u001b[0m ", log.LstdFlags|log.Lshortfile)
)

func LogInfo(msg string, v ...interface{}) {
	Info.Printf(msg, v...)
}

func LogError(err error, msg string) {
	Error.Printf("%s: %v", msg, err)
}