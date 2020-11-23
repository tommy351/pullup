package collection

import (
	"reflect"
)

type NamedArray struct {
	arr  []interface{}
	keys map[string]int
}

func NewNamedArray(data interface{}) (*NamedArray, bool) {
	v := reflect.ValueOf(data)

	switch v.Kind() {
	case reflect.Array, reflect.Slice:
		output := &NamedArray{
			arr:  make([]interface{}, v.Len()),
			keys: make(map[string]int, v.Len()),
		}

		for i := 0; i < v.Len(); i++ {
			name, ok := tryGetName(v.Index(i))

			if !ok {
				return nil, false
			}

			output.arr[i] = v.Index(i).Interface()
			output.keys[name] = i
		}

		return output, true

	default:
		return nil, false
	}
}

func tryGetName(v reflect.Value) (string, bool) {
	v = tryGetElem(v)

	if !v.IsValid() || v.Kind() != reflect.Map {
		return "", false
	}

	name := tryGetElem(v.MapIndex(reflect.ValueOf("name")))

	if name.IsValid() && name.Kind() == reflect.String {
		return name.String(), true
	}

	return "", false
}

func tryGetElem(v reflect.Value) reflect.Value {
	switch v.Kind() {
	case reflect.Interface, reflect.Ptr:
		return tryGetElem(v.Elem())
	default:
		return v
	}
}

func (n *NamedArray) Index(index int) (interface{}, bool) {
	if index < len(n.arr) {
		return n.arr[index], true
	}

	return nil, false
}

func (n *NamedArray) Get(key string) (interface{}, bool) {
	i, ok := n.keys[key]

	if !ok {
		return nil, false
	}

	return n.arr[i], true
}

func (n *NamedArray) Set(key string, value interface{}) {
	i, ok := n.keys[key]

	if ok {
		n.arr[i] = value
	} else {
		n.keys[key] = len(n.arr)
		n.arr = append(n.arr, value)
	}
}

func (n *NamedArray) Len() int {
	return len(n.arr)
}

func (n *NamedArray) Iterate() *NamedArrayIterator {
	iterator := &NamedArrayIterator{
		arr:     n.arr,
		current: -1,
		indices: map[int]string{},
	}

	for k, i := range n.keys {
		iterator.indices[i] = k
	}

	return iterator
}

func (n *NamedArray) ToArray() []interface{} {
	return n.arr
}

type NamedArrayIterator struct {
	arr     []interface{}
	current int
	indices map[int]string
}

func (n *NamedArrayIterator) Next() bool {
	n.current++

	return n.current < len(n.arr)
}

func (n *NamedArrayIterator) Key() string {
	return n.indices[n.current]
}

func (n *NamedArrayIterator) Value() interface{} {
	return n.arr[n.current]
}
