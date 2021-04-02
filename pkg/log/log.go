package log

// Logger is a logger interface.
type Logger interface {
	Print(kvpair ...interface{})
	PrintJSON(kvpair ...interface{})
}

type logger struct {
	log    Logger
	kvpair []interface{}
}

func (l *logger) Print(kvpair ...interface{}) {
	l.log.Print(append(kvpair, l.kvpair...)...)
}

func (l *logger) PrintJSON(kvpair ...interface{}) {
	l.log.PrintJSON(append(kvpair, l.kvpair...)...)
}

// With with logger kv pairs.
func With(log Logger, kvpair ...interface{}) Logger {
	return &logger{log: log, kvpair: kvpair}
}

// Debug returns a debug logger.
func Debug(log Logger) Logger {
	return With(log, LevelKey, LevelDebug)
}

// Info returns a info logger.
func Info(log Logger) Logger {
	return With(log, LevelKey, LevelInfo)
}

// Warn return a warn logger.
func Warn(log Logger) Logger {
	return With(log, LevelKey, LevelWarn)
}

// Error returns a error logger.
func Error(log Logger) Logger {
	return With(log, LevelKey, LevelError)
}
