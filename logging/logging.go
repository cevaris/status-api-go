package logging

import "github.com/cevaris/timber"

var logger = timber.NewOpLogger("status")

func Logger() timber.Logger {
	return logger
}