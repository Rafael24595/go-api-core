package translator

type DeserializeIterator struct {
	deserilizer CsvtDeserializer
	max         int
	current     int
}

func newIterator(deserilizer CsvtDeserializer, max int) DeserializeIterator {
	return DeserializeIterator{
		deserilizer: deserilizer,
		max: max,
		current: -1,
	}
}

func (i *DeserializeIterator) Next() bool {
	i.current++
	return i.current < i.max
}

func (i *DeserializeIterator) Deserialize(value any) (any, TranslateError) {
	return i.deserilizer.deserializeIndex(value, i.current)
}
