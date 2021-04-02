package log

type nopLogger struct{}

func (l *nopLogger) Print(kvpair ...interface{})     {}
func (l *nopLogger) PrintJSON(kvpair ...interface{}) {}
