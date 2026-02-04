//go:build !debug
// +build !debug

package main

func DbgPrintf(format string, dumpee ...any) {
	// NOOP
}

func DbgPrintln(a any) {
	// NOOP
}

func DbgSanitizedPrintln(message string) {
}

func DbgSanitizedPrintf(format string, dumpee string) {
}
