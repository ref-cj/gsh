package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	for {
		fmt.Fprint(os.Stdout, "$ ")

		command, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			panic(err)
		} else {
			commandFields := strings.Fields(command)
			switch commandFields[0] {
			case "exit":
				exit(0)
			case "echo":
				echo(commandFields[1:])
			default:
				fmt.Printf("%v: command not found\n", strings.Join(commandFields, ""))
			}
		}
	}
}

func echo(params []string) {
	fmt.Println(strings.Join(params, " "))
}

func exit(code int) {
	os.Exit(code)
}
