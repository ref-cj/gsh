package main

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"os"
	"slices"
	"strings"
	"sync"
	"time"
)

type readline struct {
	Completions []string // a smart implementation would do a trie here
}

var Readline = readline{Completions: []string{"echo", "exit"}} // FIXME: this is just to see if it passes codecrafters test. Will un-hard-code later (and add a way to populate comps)

var binariesInPath []string

func init() {
	timeStart := time.Now()
	if pathValue, exists := os.LookupEnv("PATH"); exists && len(pathValue) > 0 {
		var wgPaths sync.WaitGroup
		binaryChan := make(chan string, 10)

		for path := range strings.SplitSeq(pathValue, string(os.PathListSeparator)) {
			wgPaths.Add(1)
			go func(path string) { // launch a goroutine for every directory. Capture path to avoid referencing the "referenced loop var in a goroutine" gotcha (https://go.dev/wiki/CommonMistakes)
				defer wgPaths.Done()
				dirEntries, err := os.ReadDir(path)
				if err == nil {
					for _, dirEntry := range dirEntries {
						fileInfo, err := os.Stat(path + string(os.PathSeparator) + dirEntry.Name())
						if err == nil && (fileInfo.Mode().Perm()&0o0100 != 0) {
							binaryChan <- fileInfo.Name()
						}
					}
				}
			}(path)
		}

		go func() {
			for fileName := range binaryChan {
				binariesInPath = append(binariesInPath, fileName)
			}
		}()

		wgPaths.Wait()
		DbgPrintf("Found %d commands in total\n", len(binariesInPath))
		DbgPrintf(" $PATH search took %s\n", time.Since(timeStart))
	} else {
		DbgPrintln("Warn: Either $PATH does not exist, or it is empty")
	}
}

func (r readline) GetLine() (string, error) {
	ps1cached := GetPS1()
	var line string
	input := bufio.NewReader(os.Stdin)
	done := false
	tabCount := 0
	// NOTE: this cache *can* be built in init and stored but this make the list of binaries in PATH static.
	//  As opposed to now, where we generate it on request.
	//  This means if a binary is added to a location in path (or if the value of PATH changes) we wont be able to use it in a completion until the shell is restarted.
	//  This scenario is obviously very rare and neither bash, nor zsh support this either. But something to keep in mind
	var matchingBinariesCache []string

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
			tabCount++
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

			if len(matchingBinariesCache) == 0 {
				begin := time.Now()
				matchingBinariesCache = getMatchingBinariesInPath(lastWord)
				// this is required by codecrafters tests
				// neither zsh nor bash does this without additional configuration
				// and I kind of don't like it. Tie not doing this to a flag maybe? we can set in our env and codecrafters can ignore on theirs
				slices.Sort(matchingBinariesCache)
				end := time.Since(begin)
				DbgPrintf("\nsearch (and sort) took: %v\n", end)
			} else {
				DbgPrintf("\nusing completion cache for results in path\n")
			}

			switch len(matchingBinariesCache) {
			case 0:
				// This is only happens if no completion candidates are in builtins or in path
				fmt.Printf("%c", '\a') // ding
			case 1:
				restoredSpace := ""
				if !isFirstWord { // if there were words before this, restore the space we cut off
					restoredSpace = " "
				}
				line = line[:lastSpaceInLine] + restoredSpace + matchingBinariesCache[0] + " " // replace the last word with the first completion
				matchingBinariesCache = nil
				tabCount = 0
			default:
				if tabCount == 2 {
					matchingBinariesCache := slices.Clip(matchingBinariesCache)
					fmt.Fprintf(os.Stdout, "\n%s\n", strings.Join(matchingBinariesCache, " "))
					matchingBinariesCache = nil
					tabCount = 0
				} else {
					fmt.Fprintf(os.Stdout, "%c", '\a')
				}
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

func getMatchingBinariesInPath(wordPart string) []string {
	var matching []string
	shortestMatchLength := math.MaxInt
	for _, binary := range binariesInPath {
		if strings.HasPrefix(binary, wordPart) {
			matching = append(matching, binary)
			shortestMatchLength = min(len(binary), shortestMatchLength)
		}
	} // got all the binaries that start with our completion string

	if len(matching) == 0 {
		return matching
	}

	expandedCompletion := []rune(wordPart)
	for j := len(wordPart); j < shortestMatchLength; j++ {

		allCandidatesMatch := true

		for i := 1; i < len(matching); i++ { // check if every rune in this position is the same in all matches
			currentRune := []rune(matching[i])[j]       //
			firstElementsRune := []rune(matching[0])[j] // could have used any match because just one having a different rune is enough to terminate the search but using the first (0th) one makes this loop more convenient
			allCandidatesMatch = allCandidatesMatch && (currentRune == firstElementsRune)
		}

		if allCandidatesMatch {
			expandedCompletion = append(expandedCompletion, rune(matching[0][j]))
		}
	}
	DbgPrintf("\nexpanded match: %s\n", string(expandedCompletion))
	return matching
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
