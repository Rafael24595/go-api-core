package parser

type CsvDeserializer struct {
	Structures map[string][]string
}

func (s *CsvDeserializer) Deserialize(csv string) {
	tables := readTables(csv)
	println(tables)
}

func readTables(csv string) []string {
	raw := csv
	tables := []string{}
	for len(raw) > 0 {
		var table string
		table, raw = readNextTable(raw)
		if len(table) > 0 {
			tables = append(tables, table)
		}
	}
	return tables
}

func readNextTable(csv string) (string, string) {
	initial := 0
	headCount := 0

	end := 0
	tailCount := 0

	count := 0
	for i, v := range csv {
		count = i
		if v == '/' {
			if headCount == 0 {
				initial = i
			}
			headCount++
		} else if v == '\n' {
			if tailCount == 1 {
				end = i
				break
			}
			tailCount++
		} else {
			if headCount < 3 {
				headCount = 0
			}
			if tailCount < 2 {
				tailCount = 0
			}
		}
	}

	if count == len(csv) -1 {
		end = count
	}

	if end == 0 {
		return "", ""
	}

	return csv[initial:end], csv[end:]
}