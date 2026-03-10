package main

import (
	"errors"
	"fmt"
	"unicode"
)

type TokenType int

const (
	Plain TokenType = iota
	SingleQuote
	DoubleQuote
	Redirection
	Termination
)

type IToken interface {
	GetPosition() int
	GetType() TokenType
	String() string
}

type Token struct {
	Position int
	Type     TokenType
}

func (token Token) GetPosition() int {
	return token.Position
}

func (token Token) GetType() TokenType {
	return token.Type
}

func (redirectToken RedirectToken) GetPosition() int {
	return redirectToken.Token.Position
}

func (redirectToken RedirectToken) GetType() TokenType {
	return redirectToken.Token.Type
}

type RedirectType int

const (
	RedirectInput RedirectType = iota
	RedirectOutput

// RedirectError
)

type RedirectToken struct {
	Token     Token
	Direction RedirectType
}

func (t RedirectToken) String() string {
	return "TEMPORARY!!"
}

func (t Token) String() string {
	var TokenShortName string
	switch t.Type {
	case Plain:
		TokenShortName = "Pl"
	case SingleQuote:
		TokenShortName = "SQ"
	case DoubleQuote:
		TokenShortName = "DQ"
	case Redirection:
		TokenShortName = "->"
	case Termination:
		TokenShortName = "Tx"
	default:
		TokenShortName = fmt.Sprintf("?%d?", t.Type)
	}
	return fmt.Sprintf("{Pos: %d - Type: %s}", t.Position, TokenShortName)
}

func GetNextStartToken(command []rune) IToken {
	for i, r := range command {
		switch {
		case r == 10 && len(command) == 1:
			return Token{Position: i, Type: Termination}
		case r == '\'':
			return Token{Position: i, Type: SingleQuote}
		case r == '"':
			return Token{Position: i, Type: DoubleQuote}
		case (r == '>' && command[i+1] == '>') || (unicode.IsDigit(r) && command[i+1] == '>'):
			return RedirectToken{Token{i, Redirection}, RedirectOutput}

		// TODO: figure out a more generalized way of handling this
		// (or maybe we can separate path identifiers from words and numbers? so, less general? 🤔)
		case unicode.IsDigit(r), unicode.IsLetter(r), r == '/', r == '.', r == '~', r == '\\':
			return Token{Position: i, Type: Plain}
		case unicode.IsSpace(r):
			DbgPrintf("skipping something spacey: '%c' (0x%x - %d) @ position: %d\n", r, r, r, i)
			continue
		default:
			DbgPrintf("unsupported start token character %c (%d) at position %d\n", r, r, i)
			panic("unsupported beginning for a start token")
		}
	}
	panic("fell off the edge chasing a start token")
}

func GetNextPlainTokenEnd(command []rune) (Token, error) {
	DbgSanitisedPrintf("going to search in %v for Plain end token\n", string(command))
	for i := 0; i < len(command); i++ {
		r := command[i]
		if r == '\\' && command[i+1] == ' ' {
			DbgSanitisedPrintf("Escaped space in [%s]", string(command[i-1:i+2]))
			i++
		}
		if (r == '\'' || r == '"') && command[i-1] != '\\' { // if this char is un unescaped quote
			return Token{Position: i, Type: Plain}, nil
		}
		if unicode.IsSpace(r) {
			return Token{Position: i, Type: Plain}, nil
		}

		// else
		DbgPrintf("not it: %c@%d\n", r, i)
		// continue
	}
	// we chouldn't find it
	return Token{}, errors.New("fell off the edge chasing space")
}

func GetNextSingleQuoteTokenEnd(command []rune) (Token, error) {
	DbgSanitisedPrintf("going to search in %v for SingleQuote end token\n", string(command))
	// skip the first character, it will be the SingleQuote starting token
	for i := 1; i < len(command); i++ {
		r := command[i]
		if r != '\'' { // not a quote
			continue
		} else {
			return Token{Position: i + 1, Type: SingleQuote}, nil
		}
	}
	// we chouldn't find it
	return Token{}, errors.New("fell off the edge chasing single quote")
}

// TODO: we won't handle special escaping cases rn, this will be a simple dupe of GetNextSingleQuoteTokenEnd()
// (so yeah, it's intentional.. and hopefully short lived)

func GetNextDoubleQuoteTokenEnd(command []rune) (Token, error) {
	DbgSanitisedPrintf("going to search in %v for DoubleQuote end token\n", string(command))
	escaping := false
	for i := 1; i < len(command); i++ { // skip the first character, it will be the DoubleQuote starting token
		r := command[i]
		if r == '\\' {
			if escaping {
				escaping = false // consume the escape
			} else {
				DbgPrintln("will escape")
				escaping = true
			}
			continue
		}
		if r == '"' {
			if escaping {
				escaping = false // consume the escape
				DbgPrintln("used up escape")
				DbgPrintf("r: %c\n", r)
				continue
			} else {
				return Token{Position: i + 1, Type: DoubleQuote}, nil
			}
		}
		if r != '"' { // not a quote
			DbgPrintf("DQ--not it r: %c\n", r)
			continue
		}

	}
	// we chouldn't find it
	return Token{}, errors.New("fell off the edge chasing double quote")
}
