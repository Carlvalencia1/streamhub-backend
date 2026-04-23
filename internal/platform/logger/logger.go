package logger

import (
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

// Notification structured logging methods

func NotificationStreamLiveRequest(traceID, streamID, streamerID string) {
	infoLog.Printf("[NOTIFICATION] evento=stream_live_request | trace_id=%s | stream_id=%s | streamer_id=%s\n",
		traceID, streamID, streamerID)
}

func ResolveRecipients(traceID string, followerCount, streamerCount, dedupCount int) {
	infoLog.Printf("[NOTIFICATION] evento=resolve_recipients | trace_id=%s | follower_tokens=%d | streamer_tokens=%d | dedup_tokens=%d\n",
		traceID, followerCount, streamerCount, dedupCount)
}

func FCMSendBatch(traceID string, batchIndex, batchSize, successCount, failureCount int) {
	infoLog.Printf("[NOTIFICATION] evento=fcm_send_batch | trace_id=%s | batch_index=%d | batch_size=%d | success=%d | failure=%d\n",
		traceID, batchIndex, batchSize, successCount, failureCount)
}

func InvalidateTokens(traceID string, invalidCount int) {
	warnLog.Printf("[NOTIFICATION] evento=fcm_invalidate_tokens | trace_id=%s | invalid_tokens=%d\n",
		traceID, invalidCount)
}

func StreamLiveNotificationDone(traceID string, totalSent int) {
	infoLog.Printf("[NOTIFICATION] evento=stream_live_notification_done | trace_id=%s | total_sent=%d\n",
		traceID, totalSent)
}