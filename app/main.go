package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type builtin func([]string)

var builtins = make(map[string]func([]string))

// {
// 	"echo": echo,
// 	"exit": exit,
// 	"type": toipe,
// }

func main() {
	builtins["echo"] = echo
	builtins["exit"] = exit
	builtins["type"] = toipe
	for {
		fmt.Fprint(os.Stdout, "$ ")

		command, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			panic(err)
		} else {
			commandFields := strings.Fields(command)
			if builtin, ok := builtins[commandFields[0]]; ok {
				builtin(commandFields[1:])
			} else {
				fmt.Printf("%v: command not found\n", strings.Join(commandFields, ""))
			}
		}
	}
}

func echo(params []string) {
	fmt.Println(strings.Join(params, " "))
}

func exit(code []string) {
	exitCode, err := strconv.Atoi(code[0])
	if err != nil {
		panic(err)
	} else {
		os.Exit(exitCode)
	}
}

func toipe(fns []string) {
	for _, t := range fns {
		if _, ok := builtins[t]; ok {
			fmt.Printf("%s is a builtin function\n", t)
		} else {
			fmt.Printf("%s: not found\n", t)
		}
	}
}
