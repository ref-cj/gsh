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
			continue
		}
		switch {
		case r == '\'':
			return Token{Position: i, Type: SingleQuote}
		case unicode.IsDigit(r), unicode.IsLetter(r):
			return Token{Position: i, Type: Plain}
		default:
			panic("for now")
		}
	}
	panic("hit edn?")
}

func GetNextPlainTokenEnd(command []rune) (Token, error) {
	for i, r := range command {
		if unicode.IsSpace(r) {
			return Token{Position: i, Type: Plain}, nil
		} else {
			continue
		}
	}
	// we chouldn't find it
	return Token{}, errors.New("fell off the edge chasing space")
}

func GetNextSingleQuoteTokenEnd(command []rune) (Token, error) {
	panic("not implemented")
}
