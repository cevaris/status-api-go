package report

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/cevaris/timber"
)

func NewLogger(logger timber.Logger) *Logger {
	buffer := bytes.NewBuffer(make([]byte, 0))
	writer := bufio.NewWriter(buffer)
	bufferLogger := timber.NewGoBufferLogger(writer)
	return &Logger{logger: logger, reportLogger: bufferLogger, buffer: buffer, writer: writer}
}

type Logger struct {
	logger       timber.Logger // regular user defined
	reportLogger timber.Logger // logger that writes to buffer
	buffer       *bytes.Buffer // report data
	writer       *bufio.Writer // log writer ref, needed for flushing
}

func (l *Logger) Info(ctx context.Context, m ...interface{}) {
	l.logger.Info(ctx, m...)
	l.reportLogger.Info(ctx, m...)
}

func (l *Logger) Error(ctx context.Context, m ...interface{}) {
	l.logger.Error(ctx, m...)
	l.reportLogger.Error(ctx, m...)
}

func (l *Logger) Debug(ctx context.Context, m ...interface{}) {
	l.logger.Debug(ctx, m...)
	l.reportLogger.Debug(ctx, m...)
}

// Returns collected logs
func (l *Logger) Collect() []byte {
	if err := l.writer.Flush(); err != nil {
		fmt.Println("report buff logger failed to flush")
		return nil
	}
	return l.buffer.Bytes()
}
