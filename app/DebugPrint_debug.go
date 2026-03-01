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
	sanitizedRune := runeAtPosition

	switch runeAtPosition {
	case ' ':
		sanitizedRune = SpaceChar
	case '\n':
		sanitizedRune = NewLineChar
	}

	DbgPrintln(fmt.Sprintf("%s: %v : %c", message, token, sanitizedRune))
}

func sanitizeString(message string) string {
	withoutSpace := strings.ReplaceAll(message, " ", "⍽")
	withoutNewLine := strings.ReplaceAll(withoutSpace, "\n", "⏎")
	return withoutNewLine
}

func DbgSanitizedPrintln(message string) {
	DbgPrintln(sanitizeString(message))
}

func DbgSanitizedPrintf(format string, dumpee string) {
	DbgPrintf(format, sanitizeString(dumpee))
}
