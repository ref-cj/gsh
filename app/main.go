package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	for {
		fmt.Fprint(os.Stdout, "$ ")

		command, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			panic(err)
		} else {
			commandWithoutNewLine := command[:len(command)-1]
			switch commandWithoutNewLine {
			case "exit 0": // this feels like cheating
				os.Exit(0)
			default:
				fmt.Printf("%s: command not found\n", commandWithoutNewLine)
			}
		}
	}
}
