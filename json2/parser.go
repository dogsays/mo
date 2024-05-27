package json2

import (
	"fmt"
)

type jsonParser struct {
	tokens []*token
	cursor int
}

func (p *jsonParser) currentIs(t tokenType) bool {
	if p.cursor >= len(p.tokens) {
		return false
	}

	return p.tokens[p.cursor].Type == t
}

func (p *jsonParser) pick() *token {
	return p.tokens[p.cursor]
}

func (p *jsonParser) consume(t tokenType) *token {
	if p.cursor >= len(p.tokens) {
		return nil
	}

	token := p.tokens[p.cursor]

	if token.Type == t {
		p.cursor++
		return token
	}

	return nil
}

func (p *jsonParser) parseObject() (*jsonValue, error) {
	if p.consume(tokenTypeBlockStart) == nil {

		return nil, p.makeError("expected token '{'")
	}

	mp := map[string]*jsonValue{}

	for {
		if p.currentIs(tokenTypeBlockEnd) {
			break
		}

		key := p.consume(tokenTypeString)
		if key == nil {
			return nil, p.makeError("unexpected token")
		}

		if p.consume(tokenTypeColumn) == nil {
			return nil, p.makeError("unexpected token")
		}

		val, err := p.parseValue()
		if err != nil {
			return nil, err
		}

		mp[string(key.Slice())] = val

		if p.consume(tokenTypeComma) == nil {
			break
		}
	}

	if p.consume(tokenTypeBlockEnd) == nil {
		return nil, p.makeError("expected token '}'")
	}

	return &jsonValue{t: valueTypeObject, v: mp}, nil
}

func (p *jsonParser) parseArray() (*jsonValue, error) {
	if p.consume(tokenTypeArrayStart) == nil {
		return nil, p.makeError("expected token '['")
	}

	arr := []*jsonValue{}

	for {
		if p.currentIs(tokenTypeArrayEnd) {
			break
		}

		val, err := p.parseValue()
		if err != nil {
			return nil, err
		}

		arr = append(arr, val)

		if p.consume(tokenTypeComma) == nil {
			break
		}
	}

	if p.consume(tokenTypeArrayEnd) == nil {
		return nil, p.makeError("expected token ']'")
	}

	return &jsonValue{t: valueTypeArray, v: arr}, nil
}

func (p *jsonParser) parseValue() (*jsonValue, error) {
	token := p.pick()

	switch token.Type {
	case tokenTypeNull:
		p.consume(token.Type)
		return &jsonValue{t: valueTypeNull, v: token.Slice()}, nil

	case tokenTypeFloat:
		p.consume(token.Type)
		return &jsonValue{t: valueTypeFloat, v: token.Slice()}, nil

	case tokenTypeInt:
		p.consume(token.Type)
		return &jsonValue{t: valueTypeInt, v: token.Slice()}, nil

	case tokenTypeBoolean:
		p.consume(token.Type)
		return &jsonValue{t: valueTypeBool, v: token.Slice()}, nil

	case tokenTypeString:
		p.consume(token.Type)
		return &jsonValue{t: valueTypeString, v: token.Slice()}, nil

	case tokenTypeBlockStart:
		return p.parseObject()

	case tokenTypeArrayStart:
		return p.parseArray()
	}

	return nil, p.makeError("unexpected token")
}

func (p *jsonParser) makeError(s string) error {
	var t *token
	if p.cursor < len(p.tokens) {
		t = p.pick()
	} else {
		t = p.tokens[len(p.tokens)-1]
	}
	return fmt.Errorf("%v at line:%v pos:%v %v", s, t.Line, t.Start, string(t.Slice()))
}
