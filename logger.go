package smartcrop

type logger interface {
	Output(calldepth int, s string) error
}

// LogLevel specifies the severity of a given log message
type LogLevel int

// Log levels
const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarning
	LogLevelError
)

// String returns the string form for a given LogLevel
func (lvl LogLevel) String() string {
	switch lvl {
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarning:
		return "WARN"
	case LogLevelError:
		return "ERROR"
	}
	return "DEBUG"
}
