package log

import (
	"strings"

	"github.com/Rafael24595/go-api-core/src/commons/utils"
)

type Log interface {
	Name() string
	Metadata() string
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
	log := newloggerModule()
	log.pushModule(newModuleConsole())
	log.Messagef("Default logging is configured to use the %s instance. %s", log.Name(), log.Metadata())
	return log
}

func ConfigureLog(session string, timestamp int64, kargs map[string]utils.Any) Log {
	code, ok := kargs["GO_LOG_INSTANCE"]
	if !ok {
		return log
	}

	if codeStr, _ := code.String(); codeStr == CODE_LOGGER_MODULE {
		log = instanceModuleLogger(session, timestamp, kargs)
	}

	return log
}

func instanceModuleLogger(session string, timestamp int64, kargs map[string]utils.Any) Log {
	modulesInterface, ok := kargs["GO_LOG_MODULES"]
	if !ok {
		return log
	}

	modulesStr, _ := modulesInterface.String()
	modules := strings.Split(modulesStr, "|")

	newInstance := newloggerModule()
	loaded := make(map[string]int)

	for _, v := range modules {
		switch v {
		case MODULE_FILE:
			if _, ok := loaded[v]; ok {
				continue
			}
			newInstance.pushModule(newModuleFile(session, timestamp))
			loaded[MODULE_FILE] = 1
		case MODULE_CONSOLE:
			if _, ok := loaded[v]; ok {
				continue
			}
			newInstance.pushModule(newModuleConsole())
			loaded[MODULE_CONSOLE] = 1
		default:
			log.Panicf("The logger module %s is not found", v)
		}
	}

	newInstance.Messagef("The logging is configured to use the %s instance. %s", log.Name(), log.Metadata())

	return newInstance
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
