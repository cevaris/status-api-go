package logging

import (
	"github.com/cevaris/timber"
	"sync"
)

var _logger = timber.NewOpLogger("status")
var _logMap sync.Map

// STDOUT caches file based logger
func Logger() timber.Logger{
	return _logger
}

// FileLogger caches file based loggers
func FileLogger(name string) timber.Logger {
	var log timber.Logger

	if v, ok := _logMap.Load(name); ok {
		log = v.(timber.Logger)
	} else {
		log = timber.NewOpFileLogger(name)
		_logMap.Store(name, log)
	}

	return log
}
