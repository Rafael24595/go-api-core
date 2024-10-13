package csvt_translator

import (
	"fmt"
	"reflect"
)

type CsvtDeserializer struct {
	tables ResourceCollection
}

func NewDeserialzer(csv string) (*CsvtDeserializer, TranslateError) {
	tables, err := newDeserializerReader().
		read(csv)
	if err != nil {
		return nil, err
	}
	return &CsvtDeserializer{
		tables: *tables,
	}, nil
}

func (d *CsvtDeserializer) Deserialize(value any) (any, TranslateError) {
	return d.deserializeIndex(value, 0)
}

func (d *CsvtDeserializer) deserializeIndex(value any, index int) (any, TranslateError) {
	valPtr := reflect.ValueOf(value)

	if valPtr.Kind() != reflect.Ptr || valPtr.Elem().Kind() != reflect.Struct {
		return nil, TranslateErrorFrom("Root struct must be a pointer.")
	}

	root, ok := d.tables.root()
	if !ok {
		return nil, TranslateErrorFrom("Root struct is not defined.")
	}

	group, ok := root.get(index)
	if !ok {
		return nil, TranslateErrorFrom("Index does not exists.")
	}

	result, err := d.makeElement(value, group)
	if err != nil {
		return nil, err
	}
	
	return result.Interface(), nil
}

func (d *CsvtDeserializer) Iterate() DeserializeIterator {
	max := 0
	if root, ok := d.tables.root(); ok {
		max = root.nodes.Size()
	}
	return newIterator(*d, max)
}

func (d *CsvtDeserializer) makeElement(template any, root *ResourceGroup) (reflect.Value, TranslateError) {
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

func (d *CsvtDeserializer) makeStr(template any, root *ResourceGroup) (reflect.Value, TranslateError) {
	structure := fixStr(template)

	for i := 0; i < structure.NumField(); i++ {
		name := structure.Type().Field(i).Name
		field := structure.FieldByName(name)
		fieldTemplate := field.Interface()

		node, ok := root.findField(name)
		if !ok {
			return reflect.Value{}, TranslateErrorFrom(fmt.Sprintf("Field \"%s\" not found.", name))
		}

		if !isCommonType(fieldTemplate) {
			reference, ok := d.tables.Find(node)
			if !ok {
				return reflect.Value{}, TranslateErrorFrom(fmt.Sprintf("Field \"%s\" reference \"%s\" not found.", name, node.key()))
			}
			element, err := d.makeElement(fieldTemplate, reference)
			if err != nil {
				return reflect.Value{}, err
			}
			field.Set(element)

			continue
		}

		if !field.IsValid() {
			return reflect.Value{}, TranslateErrorFrom(fmt.Sprintf("Field \"%s\" is not valid.", name))
		}
		if !field.CanSet() {
			return reflect.Value{}, TranslateErrorFrom(fmt.Sprintf("Field \"%s\" cannot set.", name))
		}

		valueRef := reflect.ValueOf(node.value)
		if field.Type() != valueRef.Type() {
			if valueRef.Type().ConvertibleTo(field.Type()) {
				valueRef = valueRef.Convert(field.Type())
			} else {
				err := fmt.Sprintf("Field \"%s\" type must be \"%s\", but \"%s\" found.", name, field.Type().Name(), valueRef.Type().Name())
				return reflect.Value{}, TranslateErrorFrom(err)
			}
		}

		field.Set(valueRef)
	}
	return structure, nil
}

func fixStr(value any) reflect.Value {
	element := reflect.ValueOf(value)
	if element.Kind() != reflect.Ptr {
		structureType := reflect.TypeOf(value)
		return reflect.New(structureType).Elem()
	}
	return element.Elem()
}

func (d *CsvtDeserializer) makeMap(template any, root *ResourceGroup) (reflect.Value, TranslateError) {
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
				return reflect.Value{}, TranslateErrorFrom(fmt.Sprintf("Field \"%s\" is not valid.", k))
			}

			v, err := d.makeElement(mapValuesElement.Interface(), reference)
			if err != nil {
				return reflect.Value{}, err
			}

			mapp.SetMapIndex(index.Convert(mapKeysType), v)
		}
	}
	return mapp, nil
}

func (d *CsvtDeserializer) makeArr(template any, root *ResourceGroup) (reflect.Value, TranslateError) {
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
			if elem.Type() != arrValuesType && !elem.CanConvert(arrValuesType) {
				err := fmt.Sprintf("Array position \"%d\" type must be \"%s\", but \"%s\" found.", i, elem.Type().Name(), arrValuesType)
				return reflect.Value{}, TranslateErrorFrom(err)
			}

			arr.Index(i).Set(elem.Convert(arrValuesType))
		} else {
			reference, ok := d.tables.Find(&v)
			if !ok {
				return reflect.Value{}, TranslateErrorFrom(fmt.Sprintf("Array position \"%d\" reference \"%s\" not found.", i, v.key()))
			}

			v, err := d.makeElement(arrValuesElement.Interface(), reference)
			if err != nil {
				return reflect.Value{}, err
			}
			arr.Index(i).Set(v.Convert(arrValuesType.Elem()))
		}
	}
	return arr, nil
}

func makeObj(template any, root *ResourceGroup) (reflect.Value, TranslateError) {
	element := reflect.ValueOf(template)

	node, ok := root.findValue()
	if !ok {
		return reflect.Value{}, TranslateErrorFrom(fmt.Sprintf("Field category \"%s\" not found.", root.category))
	}


	valueRef := reflect.ValueOf(node.value)
	if element.Type().Kind() != valueRef.Type().Kind() {
		err := fmt.Sprintf("Field category \"%s\" type must be \"%s\", but \"%s\" found.", root.category, element.Type().Name(), valueRef.Type().Name())
		return reflect.Value{}, TranslateErrorFrom(err)
	}

	return valueRef.Convert(element.Type()), nil
}
