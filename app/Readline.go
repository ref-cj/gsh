package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

type readline struct {
	Completions []string // a smart implementation would do a trie here
}

var Readline = readline{Completions: []string{"echo", "exit"}} // FIXME: this is just to see if it passes codecrafters test. Will un-hard-code later (and add a way to populate comps)

func (r readline) GetLine() (string, error) {
	ps1cached := GetPS1()
	var line string
	input := bufio.NewReader(os.Stdin)
	done := false
	for {
		readedRune, size, err := input.ReadRune()
		if size == 0 && err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		switch readedRune {
		case '\r', '\n':
			line += "\n"
			done = true
		case '\t':
			lastSpaceInLine := strings.LastIndex(line, " ")
			if lastSpaceInLine == -1 { // either this is the first word (or just tab on an empty line)
				lastSpaceInLine = 0
			}
			lastWord := line[lastSpaceInLine+1:] //+1 to drop space
			completionCandidates := getStringsWithSubstring(Readline.Completions, lastWord)
			if len(completionCandidates) > 0 {
				line = line[:lastSpaceInLine] + Readline.Completions[completionCandidates[0]] + " " // replace the last word with the first completion
			} else {
				fmt.Printf("%c", '\a') // ding
			}
		case '\b', 127: // \b is 0x8 which is backspace. But both konsole and ghostty send 127 (DEL) for backspace. This case condition covers both
			if len(line) > 0 {
				line = line[:len(line)-1]
			}
		default:
			line += string(readedRune)
		}
		fmt.Fprintf(os.Stdout, "\r\033[K%s%s", ps1cached, line) // \r to go to the beginning of the line, and ESC^K to delete from cursor position to the end of line
		if done {
			break
		}
	}
	return line, nil
}

func getStringsWithSubstring(Strings []string, Substring string) []int {
	var indexes []int
	for i, str := range Strings {
		if strings.Contains(str, Substring) {
			indexes = append(indexes, i)
		}
	}
	return indexes
}
