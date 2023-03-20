package main

import (
	"fmt"
	"io/ioutil"
	"os"
)

type Definition struct {
	Condition string `json:"condition"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: honeycomb-linter <filename>")
		os.Exit(1)
	}

	definitionFile := os.Args[1]

	definition, err := ioutil.ReadFile(definitionFile)
	if err != nil {
		fmt.Println("Error reading file:", err)
		os.Exit(1)
	}

	// Check if the condition is valid for "definition"
	_, err = ParseCondition(string(definition))
	if err != nil {
		fmt.Printf("Invalid derived column definition in file %s:\n%s\n", definitionFile, err)
		os.Exit(1)
	}

	fmt.Println("Definition is valid!")
}

type Token int

const (
	ILLEGAL Token = iota
	EOF
	WHITESPACE
	AND
	OR
	NOT
	LPAREN
	RPAREN
	EQUALS
	NOT_EQUALS
	REG_MATCH
	EXISTS
	IN
	LT
	LTE
	GT
	GTE
)

var keywords = map[string]Token{
	"AND":    AND,
	"OR":     OR,
	"NOT":    NOT,
	"=":      EQUALS,
	"!=":     NOT_EQUALS,
	"=~":     REG_MATCH,
	"EXISTS": EXISTS,
	"IN":     IN,
	"<":      LT,
	"<=":     LTE,
	">":      GT,
	">=":     GTE,
}

type Lexer struct {
	input string
	pos   int
}

func NewLexer(input string) *Lexer {
	return &Lexer{input: input}
}

func (l *Lexer) NextToken() Token {
	l.skipWhitespace()

	if l.pos >= len(l.input) {
		return EOF
	}

	ch := l.input[l.pos]
	l.pos++

	switch ch {
	case '(':
		return LPAREN
	case ')':
		return RPAREN
	case '=':
		if l.peek() == '~' {
			l.pos++
			return REG_MATCH
		}
		return EQUALS
	case '!':
		if l.peek() == '=' {
			l.pos++
			return NOT_EQUALS
		}
		return NOT
	case ',':
		return WHITESPACE
	case '<':
		if l.peek() == '=' {
			l.pos++
			return LTE
		}
		return LT
	case '>':
		if l.peek() == '=' {
			l.pos++
			return GTE
		}
		return GT
	default:
		if isLetter(ch) {
			return l.readKeyword()
		}
		return ILLEGAL
	}
}

func (l *Lexer) skipWhitespace() {
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if ch != ' ' && ch != '\t' && ch != '\r' && ch != '\n' {
			break
		}
		l.pos++
	}
}

func (l *Lexer) peek() byte {
	if l.pos >= len(l.input) {
		return 0
	}
	return l.input[l.pos]
}

func (l *Lexer) readKeyword() Token {
	start := l.pos - 1
	for l.pos < len(l.input) && isLetter(l.input[l.pos]) {
		l.pos++
	}
	word := l.input[start:l.pos]

	if tok, ok := keywords[word]; ok {
		return tok
	}
	fmt.Println(word)

	return ILLEGAL
}

func (l *Lexer) readIdentifier() string {
	start := l.pos - 1
	for l.pos < len(l.input) && isLetter(l.input[l.pos]) {
		l.pos++
	}
	return l.input[start:l.pos]
}

func (l *Lexer) readLiteral() string {
	start := l.pos - 1
	for l.pos < len(l.input) && (isLetter(l.input[l.pos]) || isDigit(l.input[l.pos])) {
		l.pos++
	}
	return l.input[start:l.pos]
}

func isWhitespace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\r' || ch == '\n'
}

func isOperator(ch byte) bool {
	return ch == '(' || ch == ')' || ch == ',' || ch == '=' || ch == '!' || ch == '<' || ch == '>'
}

func isLetter(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func ParseCondition(input string) (string, error) {
	l := NewLexer(input)

	var tokens []Token
	for {
		token := l.NextToken()

		if token == EOF {
			break
		}
		tokens = append(tokens, token)
	}

	// Check that the tokens form a valid condition
	if len(tokens) == 0 {
		return "", fmt.Errorf("Empty condition")
	}

	if tokens[0] == NOT {
		if len(tokens) == 1 {
			return "", fmt.Errorf("NOT operator must be followed by a condition")
		}
		if tokens[1] == LPAREN {
			if tokens[len(tokens)-1] != RPAREN {
				return "", fmt.Errorf("Mismatched parentheses")
			}
		}
	} else if tokens[0] == LPAREN {
		if tokens[len(tokens)-1] != RPAREN {
			return "", fmt.Errorf("Mismatched parentheses")
		}
	} else if tokens[0] == EXISTS {
		if len(tokens) == 1 {
			return "", fmt.Errorf("EXISTS operator must be followed by a field name")
		}
		if tokens[1] != WHITESPACE {
			return "", fmt.Errorf("EXISTS operator must be followed by a field name")
		}
	} else if tokens[0] == IN {
		if len(tokens) == 1 {
			return "", fmt.Errorf("IN operator must be followed by a field name")
		}
		if tokens[1] != WHITESPACE {
			return "", fmt.Errorf("IN operator must be followed by a field name")
		}
	} else if tokens[0] == REG_MATCH {
		if len(tokens) == 1 {
			return "", fmt.Errorf("=~ operator must be followed by a field name")
		}
		if tokens[1] != WHITESPACE {
			return "", fmt.Errorf("=~ operator must be followed by a field name")
		}
	} else if tokens[0] == ILLEGAL {
		return "", fmt.Errorf("Invalid condition")
	}

	// Check that the tokens form a valid condition
	if len(tokens) == 0 {
		return "", fmt.Errorf("Empty condition")
	}

	return input, nil
}
