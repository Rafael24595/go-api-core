package log

const MODULE_CONSOLE = "CONSOLE"

type moduleConsole struct {
	formatter Formatter
}

func newModuleConsole() *moduleConsole {
	return &moduleConsole{
		formatter: Formatter{},
	}
}

func (m *moduleConsole) Name() string {
	return MODULE_CONSOLE
}

func (m *moduleConsole) Vector(records []Record) []Record {
	return records
}

func (m *moduleConsole) Record(record *Record, throwPanic bool) *Record {
	message := m.formatter.Format(*record)
	println(message)
	if throwPanic {
		panic(message)
	}
	return record
}
