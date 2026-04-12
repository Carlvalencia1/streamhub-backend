package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

var (
	infoLog  = log.New(os.Stdout, "[INFO] ", log.LstdFlags)
	errorLog = log.New(os.Stderr, "[ERROR] ", log.LstdFlags)
	warnLog  = log.New(os.Stdout, "[WARN] ", log.LstdFlags)
)

func Info(msg string) {
	infoLog.Println(msg)
}

func Error(msg string) {
	errorLog.Println(msg)
}

func Warn(msg string) {
	warnLog.Println(msg)
}

func InfoWithContext(ctx string, msg string) {
	infoLog.Printf("[%s] %s\n", ctx, msg)
}

func ErrorWithContext(ctx string, msg string, err error) {
	if err != nil {
		errorLog.Printf("[%s] %s: %v\n", ctx, msg, err)
	} else {
		errorLog.Printf("[%s] %s\n", ctx, msg)
	}
}

func StreamEvent(eventType string, streamID string, details string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	infoLog.Printf("[STREAM_EVENT] %s | StreamID: %s | Type: %s | Details: %s\n", timestamp, streamID, eventType, details)
}

func Debug(msg string) {
	infoLog.Printf("[DEBUG] %s\n", msg)
}