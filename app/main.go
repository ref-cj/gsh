package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	fmt.Fprint(os.Stdout, "$ ")

	command, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		panic(err)
	} else {
		commandWithoutNewLine := command[:len(command)-1]
		fmt.Printf("%s: command not found", commandWithoutNewLine)
	}
}
