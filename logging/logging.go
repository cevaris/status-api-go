package logging

import (
	"github.com/cevaris/timber"
)

var logger = timber.NewOpLogger("status")
var logMap = make(map[string]timber.Logger)

// STDOUT caches file based logger
func Logger() timber.Logger {
	return logger
}

// FileLogger caches file based loggers
func FileLogger(name string) timber.Logger {
	var log timber.Logger

	if v, ok := logMap[name]; ok {
		log = v
	} else {
		log = timber.NewGoFileLogger(name)
		logMap[name] = log
	}

	return log
}
