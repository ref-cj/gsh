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
	var line string
	input := bufio.NewReader(os.Stdin)
parseloop:
	for {
		readedRune, _, err := input.ReadRune()
		if err != nil {
			return "", err
		}
		switch readedRune {
		case '\n':
			line += "\n"
			DbgPrintf("finished parsing line: %s\n", line)
			break parseloop
		case '\t':
			DbgPrintln("completion will go here")
		case '\b', 127: // \b is 0x8 which is backspace. But both konsole and ghostty send 127 (DEL) for backspace. This case condition covers both
			if len(line) > 0 {
				line = line[:len(line)-1]
			}
		default:
			line += string(readedRune)
		}
	}
	return line, nil
}
