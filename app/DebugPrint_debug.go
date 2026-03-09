//go:build debug
// +build debug

package main

import (
	"fmt"
	"strings"
)

func DbgPrintf(format string, dumpee ...any) {
	fmt.Printf(format, dumpee...)
}

func DbgPrintln(a any) {
	fmt.Println(a)
}

func DbgPrintTokenln(message string, token Token, runeAtPosition rune) {
	sanitisedRune := runeAtPosition

	switch runeAtPosition {
	case ' ':
		sanitisedRune = SpaceChar
	case '\n':
		sanitisedRune = NewLineChar
	}

	DbgPrintln(fmt.Sprintf("%s: %v : %c", message, token, sanitisedRune))
}

func sanitiseString(message string) string {
	withoutSpace := strings.ReplaceAll(message, " ", "⍽")
	withoutNewLine := strings.ReplaceAll(withoutSpace, "\n", "⏎")
	return withoutNewLine
}

func DbgSanitisedPrintln(message string) {
	DbgPrintln(sanitiseString(message))
}

func DbgSanitisedPrintf(format string, dumpee string) {
	DbgPrintf(format, sanitiseString(dumpee))
}
