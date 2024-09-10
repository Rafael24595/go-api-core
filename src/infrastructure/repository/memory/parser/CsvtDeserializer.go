package parser

import (
	"reflect"
)

type CsvtDeserializer struct {
	tables ResourceCollection
}

func NewDeserialzer(csv string) *CsvtDeserializer {
	tables := newDeserializerReader().
		read(csv)
	return &CsvtDeserializer{
		tables: tables,
	}
}

func (d *CsvtDeserializer) Deserialize(value any) any {
	return d.deserializeIndex(value, 0)
}

func (d *CsvtDeserializer) deserializeIndex(value any, index int) any {
	valPtr := reflect.ValueOf(value)

	if valPtr.Kind() != reflect.Ptr || valPtr.Elem().Kind() != reflect.Struct {
		panic("obj must be a pointer to a struct")
	}

	root, ok := d.tables.root()
	if !ok {
		panic("//TODO: Bad format.")
	}

	group, ok := root.get(index)
	if !ok {
		panic("//TODO: Bad format.")
	}

	result := d.makeElement(value, group)
	
	return result.Interface()
}

func (d *CsvtDeserializer) Iterate() DeserializeIterator {
	max := 0
	if root, ok := d.tables.root(); ok {
		max = root.nodes.Size()
	}
	return newIterator(*d, max)
}

func (d *CsvtDeserializer) makeElement(template any, root *ResourceGroup) reflect.Value {
	element := reflect.ValueOf(template)
	switch element.Kind() {
	case reflect.Struct, reflect.Ptr:
		return d.makeStr(template, root)
	case reflect.Map:
		return d.makeMap(template, root)
	case reflect.Slice, reflect.Array:
		return d.makeArr(template, root)
	default:
		return makeObj(template, root)
	}
}

func (d *CsvtDeserializer) makeStr(template any, root *ResourceGroup) reflect.Value {
	structure := fixStr(template)

	for i := 0; i < structure.NumField(); i++ {
		name := structure.Type().Field(i).Name
		field := structure.FieldByName(name)
		fieldTemplate := field.Interface()

		node, ok := root.findField(name)
		if !ok {
			println(5)
		}

		if !isCommonType(fieldTemplate) {
			reference, ok := d.tables.Find(node)
			if !ok {
				println(4)
			}
			element := d.makeElement(fieldTemplate, reference)
			field.Set(element)

			continue
		}

		if !field.IsValid() {
			println(1)
		}
		if !field.CanSet() {
			println(2)
		}

		valueRef := reflect.ValueOf(node.value)
		if field.Type() != valueRef.Type() {
			println(3)
		}

		field.Set(valueRef)
	}
	return structure
}

func fixStr(value any) reflect.Value {
	element := reflect.ValueOf(value)
	if element.Kind() != reflect.Ptr {
		structureType := reflect.TypeOf(value)
		return reflect.New(structureType).Elem()
	}
	return element.Elem()
}

func (d *CsvtDeserializer) makeMap(template any, root *ResourceGroup) reflect.Value {
	mapType := reflect.TypeOf(template)
	mapElement := reflect.New(mapType).Elem()
	mapKeysType := reflect.TypeOf(mapElement.Interface()).Key()
	mapValuesType := reflect.TypeOf(mapElement.Interface()).Elem()
	mapValuesElement := reflect.New(mapValuesType).Elem()

	mapp := reflect.MakeMap(mapType)

	for _, p := range root.findFields() {
		k := p.Key()
		v := p.Value()

		index := reflect.ValueOf(k)

		if v.index == -1 {
			value := reflect.ValueOf(k)
			mapp.SetMapIndex(index.Convert(mapKeysType), value.Convert(mapValuesType))
		} else {
			reference, ok := d.tables.Find(&v)
			if !ok {
				println(4)
			}

			v := d.makeElement(mapValuesElement.Interface(), reference)

			mapp.SetMapIndex(index.Convert(mapKeysType), v)
		}
	}
	return mapp
}

func (d *CsvtDeserializer) makeArr(template any, root *ResourceGroup) reflect.Value {
	arrType := reflect.TypeOf(template)
	arrElement := reflect.New(arrType).Elem()
	arrValuesType := reflect.TypeOf(arrElement.Interface()).Elem()
	arrValuesElement := reflect.New(arrValuesType).Elem()

	fields := root.findFields()
	len := len(fields)

	arr := reflect.MakeSlice(arrType, len, len)

	for i, p := range fields {
		v := p.Value()
		if v.index == -1 {
			elem := reflect.ValueOf(v.value)
			if elem.Type() != arrValuesType {
				println(3)
			}

			arr.Index(i).Set(elem.Convert(arrValuesType))
		} else {
			reference, ok := d.tables.Find(&v)
			if !ok {
				println(4)
			}

			v := d.makeElement(arrValuesElement.Interface(), reference)
			arr.Index(i).Set(v.Convert(arrValuesType.Elem()))
		}
	}
	return arr
}

func makeObj(template any, root *ResourceGroup) reflect.Value {
	element := reflect.ValueOf(template)

	node, _ := root.findValue()

	valueRef := reflect.ValueOf(node.value)
	if element.Type().Kind() != valueRef.Type().Kind() {
		println(3)
	}

	return valueRef.Convert(element.Type())
}
