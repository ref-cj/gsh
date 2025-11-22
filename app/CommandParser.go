package main

import (
	"errors"
	"unicode"
)

type TokenType int

type Token struct {
	Position int
	Type     TokenType
}

const (
	Plain TokenType = iota
	SingleQuote
)

func GetNextTokenStart(command []rune) Token {
	for i, r := range command {
		if unicode.IsSpace(r) {
			DbgPrintf("skipping something spacey: %c @ position: %d\n", r, i)
			continue
		}
		switch {
		case r == '\'':
			return Token{Position: i, Type: SingleQuote}
		// TODO: figure out a more generalized way of handling this
		case unicode.IsDigit(r), unicode.IsLetter(r), r == '/', r == '.', r == '~':
			return Token{Position: i, Type: Plain}
		default:
			DbgPrintf("eh? %c %d at position %d\n", r, r, i)
			panic("for now")
		}
	}
	panic("hit edn?")
}

func GetNextPlainTokenEnd(command []rune) (Token, error) {
	DbgPrintf("going to search in %v for Plain end token\n", string(command))
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
	for i, r := range command {
		if r == '\'' && unicode.IsSpace(command[i+1]) {
			return Token{Position: i + 1, Type: SingleQuote}, nil
		} else {
			continue
		}
	}
	// we chouldn't find it
	return Token{}, errors.New("fell off the edge chasing single quote")
}
