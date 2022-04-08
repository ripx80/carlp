package parser

import (
	"bufio"
	"bytes"
	"io"
)

const eof = rune(0)

type Token int32 // int32 is a rune

const (
	// internal tokes
	ILLEGAL Token = iota
	EOF
	WS
	EQUAL
	NEWLINE

	// Literals
	ID
	STRING
	DIGIT
	FLOAT
	ARRAY
	MAP
	TOKENSTART
	TOKENEND
)

func isWhitespace(r rune) bool { return r == ' ' || r == '\t' }
func isDigit(r rune) bool      { return (r >= '0' && r <= '9') }
func isLetter(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '_' || r == '.'
}

/*windows:\r\n | unix: \n | mac:\r macx:\n */
func isLineEnd(r rune) bool { return r == '\n' || r == '\r' }

func isTokenStart(r rune) bool { return r == '{' || r == '[' }

func isTokenEnd(r rune) bool { return r == ' ' || r == '\t' || r == ',' || r == '}' || r == ']' }

type Scanner struct {
	r *bufio.Reader
}

// NewScanner returns a new instance of Scanner.
func NewScanner(r io.Reader) *Scanner {
	return &Scanner{r: bufio.NewReader(r)}
}

// read reads the next rune from the bufferred reader.
// Returns the rune(0) if an error occurs (or io.EOF is returned).
func (s *Scanner) read() rune {
	r, _, err := s.r.ReadRune()
	if err != nil {
		return eof
	}
	return r
}

// unread places the previously read rune back on the reader.
func (s *Scanner) unread() error { return s.r.UnreadRune() }

func (s *Scanner) scanWhitespace() (tok Token, lit string) {
	var buf bytes.Buffer
	buf.WriteRune(s.read())
	for {
		if ch := s.read(); ch == eof {
			break
		} else if !isWhitespace(ch) {
			s.unread()
			break
		} else {
			buf.WriteRune(ch) // recursive
		}
	}
	return WS, buf.String()
}

// scanIdent consumes the current rune and all contiguous ident runes must start with a letter.
func (s *Scanner) scanID() (tok Token, lit string) {
	var buf bytes.Buffer
	buf.WriteRune(s.read())
	for {
		if r := s.read(); r == eof {
			break

		} else if !isLetter(r) && !isDigit(r) {
			// id can have a mix of string, numbers, _, and .
			s.unread()
			break
		} else {
			_, _ = buf.WriteRune(r) // todo handle error
		}
	}
	return ID, buf.String()
}

func (s *Scanner) scanDigit() (tok Token, lit string) {
	var buf bytes.Buffer
	buf.WriteRune(s.read())
	t := DIGIT
	for {
		r := s.read()
		if r == eof {
			break
		} else if r == '.' {
			t = FLOAT
		} else if !isDigit(r) {
			s.unread()
			break
		}
		_, _ = buf.WriteRune(r) // todo handle error
	}
	return t, buf.String()
}

func (s *Scanner) scanString() (tok Token, lit string) {
	var buf bytes.Buffer

	r := s.read()
	if r == '"' {
		//empty string detected
		return STRING, buf.String()
	}

	buf.WriteRune(r)

	for {
		if r := s.read(); r == eof {
			break
		} else if r == '"' {
			break
		} else {
			_, _ = buf.WriteRune(r)
		}
	}

	return STRING, buf.String()
}

func (s *Scanner) Scan() (tok Token, lit string) {
	r := s.read()

	if isWhitespace(r) {
		s.unread()
		return s.scanWhitespace()
	}

	if isLineEnd(r) {
		r = s.read()
		if !isLineEnd(r) {
			s.unread()
		}
		//fmt.Printf("%s\n", strconv.QuoteRune(r))

		return NEWLINE, string(r)
	}

	if isLetter(r) {
		s.unread()
		return s.scanID()
	}

	if isDigit(r) || r == '-' { // more checks for negative?
		// r = s.read() // look at next rune
		// if isDigit(r){
		// 	s.unread()
		// 	s.unread()
		// }
		s.unread()
		return s.scanDigit()
	}

	if isTokenStart(r) {
		return TOKENSTART, string(r)
	}

	if isTokenEnd(r) {
		return TOKENEND, string(r)
	}

	switch r {
	case eof:
		return EOF, ""
	case '=':
		return EQUAL, string(r)
	case '"':
		return s.scanString()
	}
	return ILLEGAL, string(r)
}
