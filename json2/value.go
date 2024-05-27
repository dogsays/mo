package json2

import (
	"encoding/json"
)

type valueType int8

const (
	valueTypeString valueType = iota
	valueTypeFloat
	valueTypeInt
	valueTypeObject
	valueTypeArray
	valueTypeBool
	valueTypeNull
)

type jsonValue struct {
	t valueType
	v any
}

func (v jsonValue) Type() valueType {
	return v.t
}

func (v jsonValue) MarshalJSON() ([]byte, error) {

	switch v.t {
	case valueTypeString:
		a := []byte{'"'}
		a = append(a, v.v.([]byte)...)
		return append(a, '"'), nil
	case valueTypeArray, valueTypeObject:
		return json.Marshal(v.v)

	default:
		return v.v.([]byte), nil
	}
}

func Unmarshal(data []byte, v any) error {

	b, err := Standardize(data)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, v)
}

func Standardize(data []byte) ([]byte, error) {

	lex := &jsonLex{buf: data}
	err := lex.Lex()
	if err != nil {
		return nil, err
	}

	p := jsonParser{tokens: lex.tokens}

	value, err := p.parseValue()
	if err != nil {
		return nil, err
	}

	return json.Marshal(value)
}

func Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}
