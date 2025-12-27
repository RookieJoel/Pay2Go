package logger
// Package logger provides structured logging
package logger

import (
	"log"
	"os"
)
































}	l.logger.Printf("[WARN] "+message, args...)func (l *Logger) Warn(message string, args ...interface{}) {// Warn logs warning level messages}	l.logger.Printf("[DEBUG] "+message, args...)func (l *Logger) Debug(message string, args ...interface{}) {// Debug logs debug level messages}	l.logger.Printf("[ERROR] "+message, args...)func (l *Logger) Error(message string, args ...interface{}) {// Error logs error level messages}	l.logger.Printf("[INFO] "+message, args...)func (l *Logger) Info(message string, args ...interface{}) {// Info logs info level messages}	}		logger: log.New(os.Stdout, "[Pay2Go] ", log.LstdFlags|log.Lshortfile),	return &Logger{func New() *Logger {// New creates a new logger}	logger *log.Loggertype Logger struct {// Logger represents application logger