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
	generalError      = 1
	commandUsageError = 2
	invalidArguments  = 3
)

type builtin func([]string)

var builtins = make(map[string]builtin)

func main() {
	builtins["echo"] = echo
	builtins["exit"] = exit
	builtins["type"] = toipe
	builtins["pwd"] = pwd
	builtins["cd"] = cd

	for {
		fmt.Fprint(os.Stdout, "\033[35m$\033[0m ")
		command, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			fmt.Println("Could not read input from stdin")
			os.Exit(commandUsageError)
		} else {
			commandFields := strings.Fields(command)
			commandName := commandFields[0]

			// is this a builtin?
			if _builtin, ok := builtins[commandName]; ok {
				_builtin(commandFields[1:])
				continue
			}

			// maybe we can run it?
			if _, exists := executableExistsInPath(commandName); exists {
				cmd := exec.Command(commandName, commandFields[1:]...)
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
	fmt.Println(strings.Join(params, " "))
}

func cd(params []string) {
	if dir, err := os.Stat(params[0]); err == nil && dir.IsDir() {
		os.Chdir(params[0])
		return
	}
	fmt.Printf("cd: %s: No such file or directory", params[0])
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
