package logger

type Logger struct {
	level Level
	e     *engine
}

// uses the [DefaultConfig], if modifying anything recommended to copy
// the default config and modify values there and pass that
//
// Do not pass this as a pointer,
// this already holds a pointer to the underlying logging engine,
// so just use this as is.
func NewLogger(l Level) Logger {
	return Logger{
		l,
		newEngine(DefaultConfig),
	}
}

func NewLoggerWithConfig(l Level, c Config) Logger {
	return Logger{
		l,
		newEngine(c),
	}
}

func (l Logger) Trace(m string, kv ...string) {
	l.e.log(levelTrace, m, kv)
}

func (l Logger) Debug(m string, kv ...string) {
	if l.level <= 1 {
		l.e.log(levelDebug, m, kv)
	}
}
func (l Logger) Info(m string, kv ...string) {
	if l.level <= 2 {
		l.e.log(levelInfo, m, kv)
	}
}
func (l Logger) Warn(m string, kv ...string) {
	if l.level <= 3 {
		l.e.log(levelWarn, m, kv)
	}
}
func (l Logger) Error(m string, kv ...string) {
	if l.level <= 4 {
		l.e.log(levelError, m, kv)
	}
}

func (l Logger) Fatal(m string, kv ...string) {
	if l.level <= 5 {
		l.e.log(levelFatal, m, kv)
		if l.e.config.Others.PanicAtFatal {
			panic("fatal log")
		}
	}
}
