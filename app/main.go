package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

const (
	noError           = 0
	generalError      = 1
	commandUsageError = 2
	invalidArguments  = 3
)

type builtin func([]string)

var builtins = make(map[string]builtin, 5)

func main() {
	builtins["echo"] = echo
	builtins["exit"] = exit
	builtins["type"] = toipe
	builtins["pwd"] = pwd
	builtins["cd"] = cd

	wd, _ := os.Getwd()
	DbgPrintf("Current working directory: %s\n", wd)

	for {
		// TODO: we should have a "last command (parsing/)execution took n milliseconds metric"
		// And maybe show it in debug mode by default?
		fmt.Print(GetPS1())
		// maybe we don't delimit by \n here? Is this baking in the assumption that every line is a new inputCommand?
		inputCommand, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			fmt.Println("Could not read input from stdin")
			os.Exit(commandUsageError)
		} else {
			var outputCommandFields []string
			inputCommandRunes := []rune(inputCommand)
			outputCommandBeingBuilt := ""

			for len(inputCommand) > 0 {
				startToken := GetNextStartToken(inputCommandRunes)
				DbgPrintTokenln("our new startToken", startToken, inputCommandRunes[startToken.Position])
				var endToken Token
				inputCommandRunes = inputCommandRunes[startToken.Position:]
				inputCommand = inputCommand[startToken.Position:]

				switch startToken.Type {
				case Plain:
					endToken, err = GetNextPlainTokenEnd(inputCommandRunes)
					if err != nil {
						fmt.Printf("Error while getting Plain End Token: %s", err)
						os.Exit(generalError)
					}
					DbgPrintTokenln("our new endToken", endToken, inputCommandRunes[endToken.Position])

				case SingleQuote:
					endToken, err = GetNextSingleQuoteTokenEnd(inputCommandRunes)
					if err != nil {
						fmt.Printf("Error while getting SingleQuote End Token: %s", err)
						os.Exit(generalError)
					}
					DbgPrintTokenln("our new endToken", endToken, inputCommandRunes[endToken.Position])

				case DoubleQuote:
					endToken, err = GetNextDoubleQuoteTokenEnd(inputCommandRunes)
					if err != nil {
						fmt.Printf("Error while getting DoubleQuote End Token: %s", err)
						os.Exit(generalError)
					}
					DbgPrintTokenln("our new endToken", endToken, inputCommandRunes[endToken.Position])

				case Termination:
					endToken, err = Token{Position: 1, Type: Termination}, nil
					DbgPrintTokenln("our new endToken", endToken, inputCommandRunes[endToken.Position-1]) // termination is a special boy

				default:
					panic("unimplemented token type")
				}
				outputCommandBeingBuilt += GetSanitisedCommandSegment(inputCommand, endToken)
				inputCommand = inputCommand[endToken.Position:]
				inputCommandRunes = inputCommandRunes[endToken.Position:]
				DbgSanitizedPrintf("new command: %v\n", inputCommand)
				DbgPrintf("new commandRunes: %v\n", inputCommandRunes)

				if (endToken.Type != Termination) && (inputCommandRunes[0] == ' ' || inputCommandRunes[0] == '\n') {
					outputCommandFields = append(outputCommandFields, outputCommandBeingBuilt)
					DbgPrintf("new commandFields: %v\n", outputCommandFields)
					outputCommandBeingBuilt = ""
				}
			}

			DbgPrintf("We are done, done. Nothing else to process.\n")
			outputCommandName := outputCommandFields[0]

			// is this a builtin?
			if _builtin, ok := builtins[outputCommandName]; ok {
				_builtin(outputCommandFields[1:])
				continue
			}

			// maybe we can run it?
			if _, exists := executableExistsInPath(outputCommandName); exists {
				cmd := exec.Command(outputCommandName, outputCommandFields[1:]...)
				DbgPrintf("running command %s with args %v\n", outputCommandName, outputCommandFields[1:])
				cmd.Stdin = os.Stdin
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				cmd.Run()
				continue

			}

			// looks like we don't know what this is
			fmt.Printf("%v: command not found\n", strings.Join(outputCommandFields, ""))
		}
	}
}

func echo(params []string) {
	fmt.Println(strings.Join(params, " "))
}

func cd(params []string) {
	// not sure about special-casing this..
	if params[0] == "~" {
		if homeDir, err := os.UserHomeDir(); err != nil {
			os.Exit(generalError)
		} else {
			params[0] = homeDir
		}
	}
	if err := os.Chdir(params[0]); err != nil {
		fmt.Printf("cd: %s: No such file or directory\n", params[0])
		// I'm sad I can't use this. Tests don't like that err starts with a lowercase 'n'
		/*
			var pathError *os.PathError
			if errors.As(err, &pathError) {
				fmt.Printf("cd: %s: %s\n", pathError.Path, pathError.Err)
			} else {
				return
			}
		*/
	}
}

func pwd(params []string) {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(generalError)
	}
	fmt.Println(wd)
}

func exit(code []string) {
	if len(code) == 0 {
		os.Exit(noError)
	}
	exitCode, err := strconv.Atoi(code[0])
	if err != nil {
		fmt.Println("first (and only) argument to exit should be an integer")
		os.Exit(invalidArguments)
	} else {
		os.Exit(exitCode)
	}
}

func toipe(fns []string) {
	for _, t := range fns {
		isBuiltin, isInPath := false, false
		if _, ok := builtins[t]; ok {
			fmt.Printf("%s is a shell builtin\n", t)
			isBuiltin = true
			continue
		}
		if fullFilePath, exists := executableExistsInPath(t); exists {
			fmt.Printf("%s is %s\n", t, fullFilePath)
			isInPath = true
			continue
		}
		if !isBuiltin && !isInPath {
			fmt.Printf("%s: not found\n", t)
		}
	}
}

func executableExistsInPath(filename string) (fullFilePath string, exists bool) {
	if pathValue, exists := os.LookupEnv("PATH"); exists && len(pathValue) > 0 {
		paths := strings.SplitSeq(pathValue, string(os.PathListSeparator))
		for p := range paths {
			fullFilePath := fmt.Sprintf("%s%c%s", p, os.PathSeparator, filename)
			if fileInfo, err := os.Stat(fullFilePath); err == nil && (fileInfo.Mode().Perm()&0o0100 != 0) {
				return fullFilePath, true
			}
		}
	}
	return "", false
}
