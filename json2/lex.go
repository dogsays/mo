package json2

import (
	"bytes"
	"fmt"
)

type tokenType uint8

const (
	tokenTypeBoolean    tokenType = iota //
	tokenTypeNull                        //
	tokenTypeString                      //
	tokenTypeFloat                       //
	tokenTypeInt                         //
	tokenTypeComma                       // ,
	tokenTypeColumn                      // :
	tokenTypeArrayStart                  // [
	tokenTypeArrayEnd                    // ]
	tokenTypeBlockStart                  // {
	tokenTypeBlockEnd                    // }
)

type token struct {
	Type  tokenType
	Line  int
	Start int
	End   int
	Buf   []byte
}

func (t *token) Slice() []byte {
	return t.Buf[t.Start:t.End]
}

func isSpace(b byte) bool {
	switch b {
	case '\t', '\n', '\v', '\f', '\r', ' ':
		return true
	}

	return false
}

func isDigit(b byte) bool      { return b >= '0' && b <= '9' }
func isAlpha(b byte) bool      { return (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z') }
func isAlnum(b byte) bool      { return isAlpha(b) || isDigit(b) }
func isIdentifier(b byte) bool { return isAlnum(b) || b == '_' }

type jsonLex struct {
	buf    []byte
	cursor int
	line   int
	pos    int
	tokens []*token
	// debug  []byte
}

func (l *jsonLex) consume() byte {
	c := l.pick()
	if c == '\n' {
		l.line++
		l.pos = 0
	}
	l.pos++
	l.cursor++
	// l.debug = l.buf[l.cursor:]
	return c
}
func (l *jsonLex) pick() byte {
	return l.buf[l.cursor]
}

func (l *jsonLex) hasMore() bool {
	return l.cursor < len(l.buf)
}

func (l *jsonLex) readString() {

	l.consume() // "

	for l.hasMore() && l.pick() != '"' {
		l.consume()
	}

	if l.hasMore() {
		l.consume() // "
	}
}

func (l *jsonLex) readNumber() tokenType {
	if l.hasMore() && (isDigit(l.pick()) || l.pick() == '-') {
		l.consume()
	}

	for l.hasMore() && isDigit(l.pick()) {
		l.consume()
	}

	if !l.hasMore() {
		return tokenTypeInt
	}

	if l.pick() != '.' {
		return tokenTypeInt
	}

	l.consume()
	for l.hasMore() && isDigit(l.pick()) {
		l.consume()
	}

	return tokenTypeFloat
}

func (l *jsonLex) readIdentifier() tokenType {

	start := l.cursor

	for l.hasMore() && isIdentifier(l.pick()) {
		l.consume()
	}

	part := l.buf[start:l.cursor]

	if bytes.Equal(part, []byte("true")) {
		return tokenTypeBoolean
	}
	if bytes.Equal(part, []byte("false")) {
		return tokenTypeBoolean
	}
	if bytes.Equal(part, []byte("null")) {
		return tokenTypeNull
	}

	return tokenTypeString
}

func (l *jsonLex) readPunct() tokenType {
	punctuation := l.consume()

	switch punctuation {
	case ',':
		return tokenTypeComma
	case ':':
		return tokenTypeColumn
	case '[':
		return tokenTypeArrayStart
	case ']':
		return tokenTypeArrayEnd
	case '{':
		return tokenTypeBlockStart
	case '}':
		return tokenTypeBlockEnd
	default:
		return l.readIdentifier()
	}
}

func (l *jsonLex) nextIs(b byte) bool {
	if l.cursor < len(l.buf)-1 {
		return l.buf[l.cursor+1] == b
	}
	return false
}

func (l *jsonLex) skipWhiteSpace() {
	for l.hasMore() && isSpace(l.pick()) {
		l.consume()
	}

	//  comment
	if l.hasMore() && l.pick() == '/' {

		if l.nextIs('/') {
			for l.pick() != '\n' && l.hasMore() {
				l.consume()
			}

		} else if l.nextIs('*') {
			l.consume()
			l.consume()

			for l.hasMore() {
				if l.pick() == '*' && l.nextIs('/') {
					l.consume()
					l.consume()
					break
				}
				l.consume()
			}
		}

		l.skipWhiteSpace()
	}
}

func (l *jsonLex) readToken() error {
	l.skipWhiteSpace()
	if !l.hasMore() {
		return nil
	}

	cursor := l.cursor

	var tokenType tokenType

	switch l.pick() {
	case '"':
		l.readString()
		tokenType = tokenTypeString
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '-':
		tokenType = l.readNumber()
	case '{', '}', '[', ']', ',', ':':
		tokenType = l.readPunct()
	default:
		if isIdentifier(l.pick()) {
			tokenType = l.readIdentifier()
		} else {
			return fmt.Errorf("invalid character:%c at %v:%v", l.pick(), l.line, l.pos)
		}
	}

	end := l.cursor
	if tokenType == tokenTypeString {
		if l.buf[cursor] == '"' {
			cursor++
		}
		if l.buf[end-1] == '"' {
			end--
		}
	}

	l.tokens = append(l.tokens, &token{
		Type:  tokenType,
		Line:  l.line,
		Start: cursor,
		End:   end,
		Buf:   l.buf,
	})

	return nil
}

func (l *jsonLex) Lex() error {
	l.line = 1
	for l.hasMore() {
		err := l.readToken()
		if err != nil {
			return err
		}
	}
	return nil
}
