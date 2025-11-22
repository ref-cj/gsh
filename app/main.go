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

var InDebugMode bool = false

type builtin func([]string)

var builtins = make(map[string]builtin)

func main() {
	builtins["echo"] = echo
	builtins["exit"] = exit
	builtins["type"] = toipe
	builtins["pwd"] = pwd
	builtins["cd"] = cd

	if len(os.Args) >= 2 {
		InDebugMode = (os.Args[1] == "debug")
	}
	const prompt = "\033[35m$\033[0m "

	for {
		fmt.Print(prompt)
		command, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			fmt.Println("Could not read input from stdin")
			os.Exit(commandUsageError)
		} else {
			var commandFields []string
			commandRunes := []rune(command)

			for len(command) > 0 {
				startToken := GetNextTokenStart(commandRunes)
				DbgPrintf("our new startToken: %v - [%c - %d ]\n", startToken, commandRunes[startToken.Position], startToken.Position)
				commandRunes = commandRunes[startToken.Position:]
				command = command[startToken.Position:]
				var endToken Token
				switch startToken.Type {
				case Plain:
					endToken, err = GetNextPlainTokenEnd(commandRunes)
				case SingleQuote:
					endToken, err = GetNextSingleQuoteTokenEnd(commandRunes)
				default:
					panic("uniplemented")
				}
				DbgPrintf("our new endToken: %v - [%c - %d ]\n", endToken, commandRunes[endToken.Position], endToken.Position)
				commandFields = append(commandFields, command[:endToken.Position])
				DbgPrintf("new commandFields: %v\n", commandFields)
				command = command[endToken.Position+1:]
				commandRunes = commandRunes[endToken.Position+1:]
				DbgPrintf("new commandRunes: %v\n", commandRunes)
				DbgPrintf("new command: %v\n", strings.ReplaceAll(command, " ", "' '"))

			}
			commandName := commandFields[0]

			// is this a builtin?
			if _builtin, ok := builtins[commandName]; ok {
				_builtin(commandFields[1:])
				continue
			}

			// maybe we can run it?
			if _, exists := executableExistsInPath(commandName); exists {
				cmd := exec.Command(commandName, commandFields[1:]...)
				DbgPrintf("running command %s with args %v\n", commandName, commandFields)
				cmd.Stdin = os.Stdin
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				cmd.Run()
				continue

			}

			// looks like we don't know what this is
			fmt.Printf("%v: command not found\n", strings.Join(commandFields, ""))
		}
	}
}

func echo(params []string) {
	fmt.Println(strings.ReplaceAll(strings.Join(params, " "), "'", ""))
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
