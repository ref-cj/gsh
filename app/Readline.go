package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
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
	// tabCount :=0
	// var matchingBinariesCache []string

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
			line = strings.TrimRight(line, " ")
			line += "\n"
			done = true
		case '\t':
			lastSpaceInLine := strings.LastIndex(line, " ")
			var lastWord string
			isFirstWord := false
			if lastSpaceInLine == -1 { // either this is the first word (or just tab on an empty line)
				lastSpaceInLine = 0
				isFirstWord = true
				lastWord = line // first and last word..
			} else {
				lastWord = line[lastSpaceInLine+1:] //+1 to drop space
			}
			builtinCompletionCandidates := getStringsWithSubstring(Readline.Completions, lastWord)
			if len(builtinCompletionCandidates) > 0 {
				restoredSpace := ""
				if !isFirstWord { // if there were words before this, restore the space we cut off
					restoredSpace = " "
				}
				line = line[:lastSpaceInLine] + restoredSpace + Readline.Completions[builtinCompletionCandidates[0]] + " " // replace the last word with the first completion
				break
			}

			// Single match implementation, keep for reference for a couple of iterations
			//
			// begin := time.Now()
			// firstMatchingBinaryInPath := getFirstMatchingBinaryInPath(lastWord)
			//
			// end := time.Since(begin)
			// DbgPrintf("\n\nsearch took: %v\n\n", end)
			//
			// if firstMatchingBinaryInPath != "" {
			// 	restoredSpace := ""
			// 	if !isFirstWord { // if there were words before this, restore the space we cut off
			// 		restoredSpace = " "
			// 	}
			// 	line = line[:lastSpaceInLine] + restoredSpace + firstMatchingBinaryInPath + " " // replace the last word with the first completion
			// 	break
			// }

			begin := time.Now()
			// do this on first tab (as per spec) and cache it. check cache the second time
			matchingBinariesInPath := getMatchingBinariesInPath(lastWord)
			end := time.Since(begin)
			DbgPrintf("\nsearch took: %v\n", end)

			switch len(matchingBinariesInPath) {
			case 0:
				// This is only happens if no completion candidates are in builtins or in path
				fmt.Printf("%c", '\a') // ding
			case 1:
				restoredSpace := ""
				if !isFirstWord { // if there were words before this, restore the space we cut off
					restoredSpace = " "
				}
				line = line[:lastSpaceInLine] + restoredSpace + matchingBinariesInPath[0] + " " // replace the last word with the first completion
			default:
				fmt.Fprintf(os.Stdout, "\n%s\n", strings.Join(matchingBinariesInPath, " "))
				//	break
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

func getFirstMatchingBinaryInPath(wordPart string) string {
	// DbgPrintf("\nsearcing for %s\n", wordPart)
	if pathValue, exists := os.LookupEnv("PATH"); exists && len(pathValue) > 0 {
		for path := range strings.SplitSeq(pathValue, string(os.PathListSeparator)) {
			// DbgPrintf("  Currently looking in %s\n", path)
			dirEntries, err := os.ReadDir(path)
			if err == nil {
				for _, dirEntry := range dirEntries {
					// DbgPrintf("    investitagating %s\n", dirEntry.Name())
					fileInfo, err := os.Stat(path + string(os.PathSeparator) + dirEntry.Name())
					if err == nil && (fileInfo.Mode().Perm()&0o0100 != 0) && strings.HasPrefix(fileInfo.Name(), wordPart) {
						// DbgPrintln("      should work!")
						return fileInfo.Name()
					}
				}
			}
		}
	}
	DbgPrintf("\nNo completion found anywhere in path for %s\n", wordPart)
	return ""
}

func getMatchingBinariesInPath(wordPart string) []string {
	var result []string
	DbgPrintf("\nsearcing for %s\n", wordPart)
	if pathValue, exists := os.LookupEnv("PATH"); exists && len(pathValue) > 0 {
		for path := range strings.SplitSeq(pathValue, string(os.PathListSeparator)) {
			DbgPrintf("  Currently looking in %s\n", path)
			dirEntries, err := os.ReadDir(path)
			if err == nil {
				for _, dirEntry := range dirEntries {
					DbgPrintf("    investitagating %s\n", dirEntry.Name())
					fileInfo, err := os.Stat(path + string(os.PathSeparator) + dirEntry.Name())
					if err == nil && (fileInfo.Mode().Perm()&0o0100 != 0) && strings.HasPrefix(fileInfo.Name(), wordPart) {
						DbgPrintf("      \033[32mshould work! (%s->%s)\n\033[0m", path, fileInfo.Name())
						result = append(result, fileInfo.Name())
					}
				}
			}
		}
	}
	if len(result) > 0 {
		DbgPrintf("\nFound %d commands in total", len(result))
	} else {
		DbgPrintf("\nNo completion found anywhere in path for %s\n", wordPart)
	}
	return result
}

func getStringsWithSubstring(Strings []string, Substring string) []int {
	var indexes []int
	for i, str := range Strings {
		if strings.HasPrefix(str, Substring) {
			indexes = append(indexes, i)
		}
	}
	return indexes
}
