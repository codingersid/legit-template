package lexer

import (
	"strings"
	"unicode"
)

// TokenType represents the type of token
type TokenType int

const (
	TOKEN_TEXT TokenType = iota
	TOKEN_ECHO_ESCAPED    // {{ $var }}
	TOKEN_ECHO_RAW        // {!! $var !!}
	TOKEN_COMMENT         // {{-- comment --}}
	TOKEN_DIRECTIVE       // @directiveName
	TOKEN_DIRECTIVE_ARGS  // @directive(args)
	TOKEN_VERBATIM_START  // @verbatim
	TOKEN_VERBATIM_END    // @endverbatim
	TOKEN_EOF
)

// String returns string representation of TokenType
func (t TokenType) String() string {
	switch t {
	case TOKEN_TEXT:
		return "TEXT"
	case TOKEN_ECHO_ESCAPED:
		return "ECHO_ESCAPED"
	case TOKEN_ECHO_RAW:
		return "ECHO_RAW"
	case TOKEN_COMMENT:
		return "COMMENT"
	case TOKEN_DIRECTIVE:
		return "DIRECTIVE"
	case TOKEN_DIRECTIVE_ARGS:
		return "DIRECTIVE_ARGS"
	case TOKEN_VERBATIM_START:
		return "VERBATIM_START"
	case TOKEN_VERBATIM_END:
		return "VERBATIM_END"
	case TOKEN_EOF:
		return "EOF"
	default:
		return "UNKNOWN"
	}
}

// Position represents location in source
type Position struct {
	Line   int
	Column int
	Offset int
}

// Token represents a lexical token
type Token struct {
	Type     TokenType
	Value    string
	Args     string // For directives with arguments
	Position Position
}

// Lexer tokenizes legit template files
type Lexer struct {
	input        string
	pos          int
	line         int
	column       int
	inVerbatim   bool
	tokens       []Token
}

// New creates a new Lexer
func New(input string) *Lexer {
	return &Lexer{
		input:  input,
		pos:    0,
		line:   1,
		column: 1,
		tokens: make([]Token, 0),
	}
}

// Tokenize processes the entire input and returns all tokens
func (l *Lexer) Tokenize() ([]Token, error) {
	for l.pos < len(l.input) {
		token, err := l.nextToken()
		if err != nil {
			return nil, err
		}
		if token.Type != TOKEN_EOF {
			l.tokens = append(l.tokens, token)
		}
	}

	// Add EOF token
	l.tokens = append(l.tokens, Token{
		Type: TOKEN_EOF,
		Position: Position{
			Line:   l.line,
			Column: l.column,
			Offset: l.pos,
		},
	})

	return l.tokens, nil
}

// nextToken returns the next token from input
func (l *Lexer) nextToken() (Token, error) {
	if l.pos >= len(l.input) {
		return Token{Type: TOKEN_EOF}, nil
	}

	startPos := Position{
		Line:   l.line,
		Column: l.column,
		Offset: l.pos,
	}

	// Handle verbatim mode - everything is text until @endverbatim
	if l.inVerbatim {
		return l.scanVerbatimContent(startPos)
	}

	// Check for comment {{-- ... --}}
	if l.matchString("{{--") {
		return l.scanComment(startPos)
	}

	// Check for raw echo {!! ... !!}
	if l.matchString("{!!") {
		return l.scanRawEcho(startPos)
	}

	// Check for escaped echo {{ ... }}
	if l.matchString("{{") {
		return l.scanEscapedEcho(startPos)
	}

	// Check for escaped @ (@@) - outputs literal @
	if l.matchString("@@") {
		l.advance()
		l.advance()
		return Token{
			Type:     TOKEN_TEXT,
			Value:    "@",
			Position: startPos,
		}, nil
	}

	// Check for directive @...
	if l.current() == '@' && l.pos+1 < len(l.input) && (unicode.IsLetter(rune(l.input[l.pos+1])) || l.input[l.pos+1] == '_') {
		return l.scanDirective(startPos)
	}

	// Otherwise, it's text content
	return l.scanText(startPos)
}

// scanComment scans a comment {{-- ... --}}
func (l *Lexer) scanComment(startPos Position) (Token, error) {
	l.advanceN(4) // Skip {{--

	start := l.pos
	for l.pos < len(l.input) {
		if l.matchString("--}}") {
			content := l.input[start:l.pos]
			l.advanceN(4) // Skip --}}
			return Token{
				Type:     TOKEN_COMMENT,
				Value:    strings.TrimSpace(content),
				Position: startPos,
			}, nil
		}
		l.advance()
	}

	return Token{}, &LexerError{
		Message:  "Unclosed comment",
		Position: startPos,
	}
}

// scanRawEcho scans raw echo {!! ... !!}
func (l *Lexer) scanRawEcho(startPos Position) (Token, error) {
	l.advanceN(3) // Skip {!!
	l.skipWhitespace()

	start := l.pos
	for l.pos < len(l.input) {
		if l.matchString("!!}") {
			content := strings.TrimSpace(l.input[start:l.pos])
			l.advanceN(3) // Skip !!}
			return Token{
				Type:     TOKEN_ECHO_RAW,
				Value:    content,
				Position: startPos,
			}, nil
		}
		l.advance()
	}

	return Token{}, &LexerError{
		Message:  "Unclosed raw echo",
		Position: startPos,
	}
}

// scanEscapedEcho scans escaped echo {{ ... }}
func (l *Lexer) scanEscapedEcho(startPos Position) (Token, error) {
	l.advanceN(2) // Skip {{
	l.skipWhitespace()

	start := l.pos
	for l.pos < len(l.input) {
		if l.matchString("}}") {
			content := strings.TrimSpace(l.input[start:l.pos])
			l.advanceN(2) // Skip }}
			return Token{
				Type:     TOKEN_ECHO_ESCAPED,
				Value:    content,
				Position: startPos,
			}, nil
		}
		l.advance()
	}

	return Token{}, &LexerError{
		Message:  "Unclosed echo",
		Position: startPos,
	}
}

// scanDirective scans a directive @name or @name(args)
func (l *Lexer) scanDirective(startPos Position) (Token, error) {
	l.advance() // Skip @

	// Read directive name
	start := l.pos
	for l.pos < len(l.input) && (unicode.IsLetter(rune(l.input[l.pos])) || unicode.IsDigit(rune(l.input[l.pos])) || l.input[l.pos] == '_') {
		l.advance()
	}
	name := l.input[start:l.pos]

	// Handle @verbatim
	if name == "verbatim" {
		l.inVerbatim = true
		return Token{
			Type:     TOKEN_VERBATIM_START,
			Value:    name,
			Position: startPos,
		}, nil
	}

	// Check for arguments in parentheses
	if l.pos < len(l.input) && l.input[l.pos] == '(' {
		args, err := l.scanDirectiveArgs()
		if err != nil {
			return Token{}, err
		}
		return Token{
			Type:     TOKEN_DIRECTIVE_ARGS,
			Value:    name,
			Args:     args,
			Position: startPos,
		}, nil
	}

	return Token{
		Type:     TOKEN_DIRECTIVE,
		Value:    name,
		Position: startPos,
	}, nil
}

// scanDirectiveArgs scans directive arguments in parentheses
func (l *Lexer) scanDirectiveArgs() (string, error) {
	l.advance() // Skip (

	start := l.pos
	depth := 1
	inString := false
	stringChar := byte(0)

	for l.pos < len(l.input) && depth > 0 {
		ch := l.input[l.pos]

		// Handle string literals
		if (ch == '"' || ch == '\'') && (l.pos == 0 || l.input[l.pos-1] != '\\') {
			if !inString {
				inString = true
				stringChar = ch
			} else if ch == stringChar {
				inString = false
			}
		}

		if !inString {
			if ch == '(' {
				depth++
			} else if ch == ')' {
				depth--
			}
		}

		if depth > 0 {
			l.advance()
		}
	}

	if depth != 0 {
		return "", &LexerError{
			Message: "Unclosed parenthesis in directive arguments",
			Position: Position{
				Line:   l.line,
				Column: l.column,
				Offset: l.pos,
			},
		}
	}

	args := l.input[start:l.pos]
	l.advance() // Skip closing )

	return strings.TrimSpace(args), nil
}

// scanText scans plain text content
func (l *Lexer) scanText(startPos Position) (Token, error) {
	start := l.pos

	for l.pos < len(l.input) {
		// Stop at special sequences
		if l.matchString("{{") || l.matchString("{!!") || l.matchString("@@") {
			break
		}
		if l.current() == '@' && l.pos+1 < len(l.input) && (unicode.IsLetter(rune(l.input[l.pos+1])) || l.input[l.pos+1] == '_') {
			break
		}
		l.advance()
	}

	content := l.input[start:l.pos]
	if content == "" {
		return l.nextToken()
	}

	return Token{
		Type:     TOKEN_TEXT,
		Value:    content,
		Position: startPos,
	}, nil
}

// scanVerbatimContent scans content inside @verbatim...@endverbatim
func (l *Lexer) scanVerbatimContent(startPos Position) (Token, error) {
	start := l.pos

	for l.pos < len(l.input) {
		if l.matchString("@endverbatim") {
			content := l.input[start:l.pos]
			l.advanceN(12) // Skip @endverbatim
			l.inVerbatim = false

			if content != "" {
				// Return the content first
				l.tokens = append(l.tokens, Token{
					Type:     TOKEN_TEXT,
					Value:    content,
					Position: startPos,
				})
			}

			return Token{
				Type:     TOKEN_VERBATIM_END,
				Value:    "endverbatim",
				Position: Position{
					Line:   l.line,
					Column: l.column - 12,
					Offset: l.pos - 12,
				},
			}, nil
		}
		l.advance()
	}

	return Token{}, &LexerError{
		Message:  "Unclosed @verbatim block",
		Position: startPos,
	}
}

// Helper methods

func (l *Lexer) current() byte {
	if l.pos >= len(l.input) {
		return 0
	}
	return l.input[l.pos]
}

func (l *Lexer) advance() {
	if l.pos < len(l.input) {
		if l.input[l.pos] == '\n' {
			l.line++
			l.column = 1
		} else {
			l.column++
		}
		l.pos++
	}
}

func (l *Lexer) advanceN(n int) {
	for i := 0; i < n; i++ {
		l.advance()
	}
}

func (l *Lexer) matchString(s string) bool {
	if l.pos+len(s) > len(l.input) {
		return false
	}
	return l.input[l.pos:l.pos+len(s)] == s
}

func (l *Lexer) skipWhitespace() {
	for l.pos < len(l.input) && (l.input[l.pos] == ' ' || l.input[l.pos] == '\t') {
		l.advance()
	}
}

// LexerError represents a lexer error
type LexerError struct {
	Message  string
	Position Position
}

func (e *LexerError) Error() string {
	return e.Message
}
