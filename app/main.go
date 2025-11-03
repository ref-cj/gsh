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

func isBinaryInPath(s string) bool {
	return true
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
		if _, ok := builtins[t]; ok {
			fmt.Printf("%s is a shell builtin\n", t)
			continue
		}
		if pathValue, exists := os.LookupEnv("PATH"); exists /*&& len(pathValue) > 0*/ {
			paths := strings.Split(pathValue, string(os.PathListSeparator))
			fmt.Printf("paths: %v", paths)
			for _, p := range paths {
				if isBinaryInPath(p) {
					fmt.Printf("%s is %s%c%s", t, p, os.PathListSeparator, t)
					continue
				}
			}
			fmt.Printf("%s: not found\n", t)
		}
	}
}
