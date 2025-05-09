package log

type Log interface {
	Name() string
	Records() []Record
	Message(string) *Record
	Messagef(string, ...any) *Record
	Warning(string) *Record
	Warningf(string, ...any) *Record
	Error(error) *Record
	Errors(string) *Record
	Errorf(string, ...any) *Record
	Panic(error)
	Panics(string)
	Panicf(string, ...any)
}

var log Log = defaultLog()

func defaultLog() Log {
	log := newConsole(&Formatter{})
	log.Messagef("Default logging is configured to use the %s instance", log.Name())
	return log
}

func Name() string {
	return log.Name()
}

func Records() []Record {
	return log.Records()
}

func Message(message string) {
	log.Message(message)
}

func Messagef(format string, args ...any) {
	log.Messagef(format, args...)
}

func Warning(message string) {
	log.Warning(message)
}

func Warningf(format string, args ...any) {
	log.Warningf(format, args...)
}

func Error(err error) {
	log.Error(err)
}

func Errors(message string) {
	log.Errors(message)
}

func Errorf(format string, args ...any) {
	log.Errorf(format, args...)
}

func Panic(err error) {
	log.Panic(err)
}

func Panics(message string) {
	log.Panics(message)
}

func Panicf(format string, args ...any) {
	log.Panicf(format, args...)
}
