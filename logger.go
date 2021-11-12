package opm

type (
	Logger interface {
		Debugf(format string, args ...interface{})
		Info(i ...interface{})
		Infof(format string, args ...interface{})
		Warn(i ...interface{})
		Warnf(format string, args ...interface{})
		Error(i ...interface{})
		Errorf(format string, args ...interface{})
		Fatal(i ...interface{})
		Fatalf(format string, args ...interface{})
		Panic(i ...interface{})
		Panicf(format string, args ...interface{})
	}
)
