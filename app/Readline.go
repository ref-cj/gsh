package main

import (
	"bufio"
	"os"
)

// I don't know how I feel about this namespacing trick I invented...
// I should look into this and confirm that this with the function receiver syntax sugar does not generate dumb code.
type readline struct{}

var Readline readline

func (r readline) GetLine() (string, error) {
	// var buf [4]byte // UTF8 can be 4 bytes long right!?
	var line string
	input := bufio.NewReader(os.Stdin)
parseloop:
	for {
		readedRune, _, err := input.ReadRune()
		if err != nil {
			return "", err
			// panic("could not read input from stdin")
		}
		switch readedRune {
		case '\n':
			line += "\n"
			DbgPrintf("finished parsing line: %s\n", line)
			break parseloop
		case '\t':
			DbgPrintln("completion will go here")
		default:
			line += string(readedRune)
		}
	}
	return line, nil
}
