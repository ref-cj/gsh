package main

import (
	"fmt"
)

func DbgPrintf(format string, dumpee ...any) {
	if InDebugMode {
		fmt.Printf(format, dumpee...)
	}
	// else NOOP
}

func DbgPrintln(a any) {
	if InDebugMode {
		fmt.Println(a)
	}
	// else NOOP
}
