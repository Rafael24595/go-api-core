package parser

type DeserializeIterator struct {
	deserilizer CsvDeserializer
	max         int
	current     int
}

func newIterator(deserilizer CsvDeserializer, max int) DeserializeIterator {
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

func (i *DeserializeIterator) Deserialize(value any) any {
	return i.deserilizer.deserializeIndex(value, i.current)
}
