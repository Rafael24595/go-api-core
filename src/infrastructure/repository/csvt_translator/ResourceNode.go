package csvt_translator

import "fmt"

type ResourceNode struct {
	value     interface{}
	index     int
}

func fromPointer(value interface{}, index int) ResourceNode {
	return ResourceNode{
		value:     value,
		index:     index,
	}
}

func fromNonPointer(value interface{}) ResourceNode {
	return ResourceNode{
		value:     value,
		index:     -1,
	}
}

func fromEmpty() ResourceNode {
	return fromNonPointer("")
}

func (n ResourceNode) key() string {
	return fmt.Sprintf("%v", n.value)
}