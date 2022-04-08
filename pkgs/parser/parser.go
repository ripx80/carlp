package parser

import (
	"errors"
	"fmt"
	"io"
	"strconv"
)

const unkownKey = "unkown"

// Errors
var (
	ErrorNullValue     = errors.New("value is null")
	ErrorKeyEqualSign  = errors.New("key equal sign missing")
	ErrorInvalidStruct = errors.New("invalid or unkown structure")
	ErrorInvalidKey    = errors.New("invalid or unkown key type")
	ErrorMixedNested   = errors.New("invalid nested type, mixed map and array")
)

// Parser represents a parser.
type Parser struct {
	s   *Scanner
	buf struct {
		tok Token  // last read token
		lit string // last read literal
		n   int    // buffer size (max=1)
	}
	m        map[string]interface{} //export
	ln       uint64                 // export
	UndefKey uint64
}

type KeyVal struct {
	Key    string // only support string keys like json
	KeyTok Token
	Value  interface{}
	ValTok Token
}

// NewParser returns a new instance of Parser.
func NewParser(r io.Reader) *Parser {
	return &Parser{s: NewScanner(r), m: make(map[string]interface{}), ln: 1}
}

// scan returns the next token from the underlying scanner.
// If a token has been unscanned then read that instead.
func (p *Parser) scan() (tok Token, lit string) {
	// If we have a token on the buffer, then return it.
	if p.buf.n != 0 {
		p.buf.n = 0
		return p.buf.tok, p.buf.lit
	}
	tok, lit = p.s.Scan()
	// Save it to the buffer in case we unscan later.
	p.buf.tok, p.buf.lit = tok, lit
	return
}

// unscan pushes the previously read token back onto the buffer.
func (p *Parser) unscan() { p.buf.n = 1 } // not used at the moment

// scanIgnoreWhitespace scans the next non-whitespace token.
func (p *Parser) scanIgnoreWhitespace() (tok Token, lit string) {
	tok, lit = p.scan()
	if tok == WS {
		tok, lit = p.scan()
	}
	return
}

func (p *Parser) ParseType(tok Token, v string) (interface{}, error) {
	//fmt.Printf("tok: %d, lit: %s\n", tok, lit)
	var (
		a   []interface{}
		r   interface{}
		err error
	)
	switch tok {
	case DIGIT:
		r, err = strconv.Atoi(v)
	case FLOAT:
		r, err = strconv.ParseFloat(v, 32)
	case TOKENSTART:
		// value is nested, replace with array
		r = a
	default:
		r = v
	}
	if err != nil {
		return nil, err
	}
	return r, err
}

func (p *Parser) ParseNested() (interface{}, error) {
	var a []interface{}
	var m = make(map[string]interface{})
	var err error

	for {
		kv, err := p.ParseKeyVal()

		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		if kv.KeyTok == TOKENEND {
			break
		}

		// check if we have a token start -> nested: {}
		if kv.KeyTok == TOKENSTART {
			// Parse?
			a = append(a, kv.Value)
			continue
		}

		// check if we have a key without value -> slice
		// its an array, not support mixed
		if kv.Value == nil {
			kv.Value, err = p.ParseType(kv.KeyTok, kv.Key)
			if err != nil {
				break
			}
			a = append(a, kv.Value)
			continue
		}

		// double here and in parse
		// check duplicate keys in maps
		if _, ok := m[kv.Key].([]interface{}); ok {
			//fmt.Println("slice interface")
			// its a slice interface.
			m[kv.Key] = append([]interface{}{kv.Value}, m[kv.Key].([]interface{})...)
			continue
		}

		if _, ok := m[kv.Key].(interface{}); ok {
			//fmt.Println("normal value in map")
			m[kv.Key] = append([]interface{}{m[kv.Key]}, kv.Value)
			continue
		}
		m[kv.Key] = kv.Value // default key=value
	}

	if len(a) == 0 {
		return m, err
	}

	if len(m) > 0 {
		//fmt.Println("array:", a)
		//fmt.Println("map:", m)
		return nil, ErrorMixedNested
	}
	return a, err
}

// Parse
func (p *Parser) ParseKeyVal() (*KeyVal, error) {
	kv := &KeyVal{}
	var err error

	for {
		kv.KeyTok, kv.Key = p.scanIgnoreWhitespace()
		switch kv.KeyTok {
		case EOF:
			return nil, io.EOF
		case NEWLINE:
			p.ln++ // overflow?
			continue
		case EQUAL:
			fmt.Println("line: ", p.ln, " skip malformed equal begin")
			_, _ = p.scanIgnoreWhitespace()
			continue

		case TOKENSTART:
			//fmt.Println("tokenstart")
			kv.Value, err = p.ParseNested()
			return kv, err

		case STRING, ID, DIGIT, FLOAT, TOKENEND:
			// this are valid tokens for keys, no type cast
		default:
			fmt.Printf("unknown key line: %d, type: %d, LIT: %s\n", p.ln, kv.KeyTok, kv.Key)
			return kv, ErrorInvalidKey
		}

		tok, lit := p.scanIgnoreWhitespace()

		// lit can be "

		if tok != EQUAL && tok != TOKENSTART {
			//fmt.Printf("no equal or nested: %d, %s\n", tok, lit)
			// no equal and no nested without equal: no value found, could be something like this: test={ 12 2354545 }
			p.unscan()
			return kv, nil
		}

		if tok == EQUAL {
			// key value pair, read value
			tok, lit = p.scanIgnoreWhitespace()
			kv.Value, err = p.ParseType(tok, lit)
		}
		// value can be tokenstart without equal sign
		if tok == TOKENSTART {
			kv.Value, err = p.ParseNested()
			//fmt.Println("nested: tok:", kv.Value, &kv.Value)
			if err != nil {
				fmt.Printf("ln: %d, problematic key: %s, val: %s\n", p.ln, kv.Key, kv.Value)
				return kv, err
			}
		}
		//fmt.Printf("key: %s, val: %s\n", kv.Key, kv.Value)

		return kv, err
	}
}

func (p Parser) Parse() (map[string]interface{}, uint64, error) {
	var (
		err error
		kv  *KeyVal
	)
	for {
		kv, err = p.ParseKeyVal()
		//fmt.Println(kv.Value)

		if err != nil {
			break
		}
		// this should never happen
		if kv.KeyTok == TOKENSTART {
			p.UndefKey++
			kv.Key = unkownKey // check this
		}

		if _, ok := p.m[kv.Key]; !ok {
			p.m[kv.Key] = kv.Value // default key=value
			continue
		}

		// map[string]interface{}
		//fmt.Println("line: ", p.ln, "douplicate key:", kv.Key)

		if _, ok := p.m[kv.Key].([]interface{}); ok {
			//fmt.Println("slice interface")
			// its a slice interface.
			p.m[kv.Key] = append([]interface{}{kv.Value}, p.m[kv.Key].([]interface{})...)

			// } else if _, ok := p.m[kv.Key].(map[string]interface{}); ok {
			// 	fmt.Println("map interface")
			// 	// its a map interface
			// 	p.m[kv.Key] = append([]interface{}{kv.Value}, p.m[kv.Key].([]interface{})...)
		} else {
			//fmt.Println("something else")
			p.m[kv.Key] = append([]interface{}{p.m[kv.Key]}, kv.Value)
		}

	}

	if err != io.EOF {
		return p.m, p.ln, err
	}
	return p.m, p.ln, nil
}
