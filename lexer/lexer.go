package lexer

import (
	"fmt"
	"io"
	"strings"
	"unicode"
)

type Token int

const (
	EOF Token = iota

	NUMBER

	NEXT_CELL
	PREVIOUS_CELL

	INCREMENT_CELL
	DECREMENT_CELL

	OUTPUT_CELL
	INPUT_CELL

	BRACKET_OPEN
	BRACKET_CLOSE
)

var (
	tokensName = []string{
		EOF: "EOF",

		NUMBER: "NUMBER",

		NEXT_CELL:     ">",
		PREVIOUS_CELL: "<",

		INCREMENT_CELL: "+",
		DECREMENT_CELL: "-",

		OUTPUT_CELL: ".",
		INPUT_CELL:  ",",

		BRACKET_OPEN:  "[",
		BRACKET_CLOSE: "]",
	}

	// Use tokensName to generate a map of tokens to their string representation
	tokensIndex = func() map[string]Token {
		tokens := make(map[string]Token)

		for token, symbol := range tokensName {
			tokens[symbol] = Token(token)
		}

		return tokens
	}()
)

func (token Token) String() string {
	return tokensName[token]
}

type Position struct {
	Line   int
	Column int
}

type Lexer struct {
	Position Position
	Reader   *strings.Reader
	Debug    *bool
}

func NewLexer(input string, debug *bool) *Lexer {
	return &Lexer{
		Reader: strings.NewReader(input),
		Debug:  debug,
	}
}

// Resets the position of the cursor
func (lexer *Lexer) resetPosition() {
	lexer.Position.Line += 1
	lexer.Position.Column = 0
}

// "Unreads" (backups) the last rune read, so that it can be read again later on with a different context
func (lexer *Lexer) backup() {
	if unreadRuneError := lexer.Reader.UnreadRune(); unreadRuneError != nil {
		panic(fmt.Sprintf("Error while trying to unread rune: %s", unreadRuneError.Error()))
	}

	lexer.Position.Column -= 1
}

// Attempts to lex a special token given an unicode function
func (lexer *Lexer) lexToken(unicodeFunction func(rune) bool) string {
	var literal string

	for {
		readRune, _, readRuneError := lexer.Reader.ReadRune()
		if readRuneError != nil {
			if readRuneError == io.EOF {
				return literal
			}

			panic(fmt.Sprintf("Error while trying to read rune: %s", readRuneError.Error()))
		}

		lexer.Position.Column += 1

		if unicodeFunction(readRune) {
			literal = literal + string(readRune)
		} else {
			lexer.backup()
			return literal
		}
	}
}

// Read and scan the input for the next rune, assigning it a token if possible
func (lexer *Lexer) Lex() (Position, Token, string) {
	for {
		readRune, _, readRuneError := lexer.Reader.ReadRune()
		if readRuneError != nil {
			if readRuneError == io.EOF {
				return lexer.Position, EOF, "\n"
			}

			panic(fmt.Sprintf("Error while trying to read rune: %s", readRuneError.Error()))
		}

		// Advance the cursor position by one
		lexer.Position.Column += 1

		switch readRune {
		case '\n':
			lexer.resetPosition()
		case '#':
			lexer.lexToken(func(readRune rune) bool {
				return readRune != '\n' // Ignore everything until the end of the line
			})
		case
			'>',
			'<',
			'+',
			'-',
			'.',
			',',
			'[',
			']':
			return lexer.Position, tokensIndex[string(readRune)], string(readRune)
		default:
			if unicode.IsSpace(readRune) {
				continue
			} else if unicode.IsDigit(readRune) {
				lexer.backup()
				return lexer.Position, NUMBER, lexer.lexToken(unicode.IsDigit)
			} else if unicode.IsLetter(readRune) {
				lexer.lexToken(func(readRune rune) bool {
					// We don't want to ignore actual symbols
					if _, ok := tokensIndex[string(readRune)]; ok {
						// We don't care about possible errors or the actual rune here
						defer lexer.Reader.ReadRune()

						// Unread the last rune so that we can verify if we really want to accept this symbol
						lexer.Reader.Seek(-2, io.SeekCurrent)

						previousReadRune, _, previousReadRuneError := lexer.Reader.ReadRune()
						if previousReadRuneError != nil {
							panic(fmt.Sprintf("Error while trying to read rune: %s", previousReadRuneError.Error()))
						}

						// We only ignore if the previous read rune is a letter
						return unicode.IsLetter(previousReadRune)
					}

					return readRune != '\n' // Ignore everything until the end of the line
				})
			} else {
				continue
			}
		}
	}
}
