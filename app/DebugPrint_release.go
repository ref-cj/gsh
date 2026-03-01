//go:build !debug
// +build !debug

package main

func DbgPrintf(format string, dumpee ...any) {
	// NOOP
}

func DbgPrintln(a any) {
	// NOOP
}

func DbgPrintTokenln(message string, token Token, runeAtPosition rune) {
	// NOOP
}

func DbgSanitizedPrintln(message string) {
}

func DbgSanitizedPrintf(format string, dumpee string) {
}
