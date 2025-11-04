package main

import (
	"bufio"
	"fmt"
	"os"
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

	for {
		fmt.Fprint(os.Stdout, "$ ")
		if false {
			fmt.Println("-- debugging")
			pathEnvVar := os.Getenv("PATH")
			for p := range strings.SplitSeq(pathEnvVar, string(os.PathListSeparator)) {
				fmt.Println(p)
				if strings.HasPrefix(p, "/tmp") {
					filesInDir, _ := os.ReadDir(p)
					for file := range filesInDir {
						fmt.Println(file)
					}
				}
			}
			fmt.Println("-- /debugging")
		}
		command, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			fmt.Println("Could not read input from stdin")
			os.Exit(commandUsageError)
		} else {
			commandFields := strings.Fields(command)

			// is this a builtin?
			if _builtin, ok := builtins[commandFields[0]]; ok {
				_builtin(commandFields[1:])
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
		if pathValue, exists := os.LookupEnv("PATH"); exists /*&& len(pathValue) > 0*/ {
			paths := strings.SplitSeq(pathValue, string(os.PathListSeparator))
			for p := range paths {
				fullFilePath := fmt.Sprintf("%s%c%s", p, os.PathSeparator, t)
				if fileInfo, err := os.Stat(fullFilePath); err == nil && (fileInfo.Mode().Perm()&0o0100 != 0) {
					fmt.Printf("%s is %s\n", t, fullFilePath)
					isInPath = true
					break
				}
			}
			if !isBuiltin && !isInPath {
				fmt.Printf("%s: not found\n", t)
			}
		}
	}
}
