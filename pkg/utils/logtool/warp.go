package logtool

import (
	"github.com/go-logr/logr"
	"time"
)

func SpendTimeRecord(log logr.Logger, message string) func() {
	startAt := time.Now()
	return func() {
		endAt := time.Now()
		log.Info(message, "cost", endAt.Sub(startAt).String())
	}
}
