package logger

import (
	"log"
)

// Logger contains a logger.
type Logger struct {
	DebugMode bool
	Log       *log.Logger
}
