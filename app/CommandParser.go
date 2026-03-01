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
	Termination
)

type Token struct {
	Position int
	Type     TokenType
}

func (t Token) String() string {
	var TokenShortName string
	switch t.Type {
	case Plain:
		TokenShortName = "Pl"
	case SingleQuote:
		TokenShortName = "SQ"
	case Termination:
		TokenShortName = "Tx"
	default:
		TokenShortName = fmt.Sprintf("?%d?", t)
	}
	return fmt.Sprintf("{Pos: %d - Type: %s}", t.Position, TokenShortName)
}

func GetNextStartToken(command []rune) Token {
	for i, r := range command {
		switch {
		case r == 10 && len(command) == 1:
			return Token{Position: i, Type: Termination}
		case r == '\'':
			return Token{Position: i, Type: SingleQuote}
		// TODO: figure out a more generalized way of handling this
		// (or maybe we can separate path identifiers from words and numbers? so, less general? 🤔)
		case unicode.IsDigit(r), unicode.IsLetter(r), r == '/', r == '.', r == '~':
			return Token{Position: i, Type: Plain}
		case unicode.IsSpace(r):
			DbgPrintf("skipping something spacey: %c @ position: %d\n", r, i)
			continue
		default:
			DbgPrintf("unsupported start token character %c (%d) at position %d\n", r, r, i)
			panic("unsupported beginning for a start token")
		}
	}
	panic("fell off the edge chasing a start token")
}

func GetNextPlainTokenEnd(command []rune) (Token, error) {
	DbgSanitizedPrintf("going to search in %v for Plain end token\n", string(command))
	for i, r := range command {
		if unicode.IsSpace(r) {
			return Token{Position: i, Type: Plain}, nil
		} else {
			DbgPrintf("not it: %c@%d\n", r, i)
			continue
		}
	}
	// we chouldn't find it
	return Token{}, errors.New("fell off the edge chasing space")
}

func GetNextSingleQuoteTokenEnd(command []rune) (Token, error) {
	DbgSanitizedPrintf("going to search in %v for SingleQuote end token\n", string(command))
	// skip the first character, it will be the SingleQuote starting token
	for i := 1; i < len(command); i++ {
		r := command[i]
		if r != '\'' { // not a quote
			continue
		} else { // we found a single quote.. Check if we should stop searching
			if i < len(command) && command[i+1] == '\'' {
				// we are not at the end, and the next char is also a single quote
				// meaning we are in "consecutive quoted strings" territory..
				// we should just skip and keep going
				i++
				continue
			} else {
				return Token{Position: i, Type: SingleQuote}, nil
			}
		}
	}
	// we chouldn't find it
	return Token{}, errors.New("fell off the edge chasing single quote")
}
