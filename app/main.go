package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// Did not use iota here because I wanted the error codes to be stable
// and possibly have a hierarcy (e.g.) 4:IOError, 41:FilePermissionError, 42:FileWriteError
// the jury is still out if it turned out to be useful :)
const (
	noError           = 0
	generalError      = 1
	commandUsageError = 2
	invalidArguments  = 3
	IOError           = 4
)

type redirections struct {
	in  *os.File
	out *os.File
	err *os.File
}

// considered generalizing this into a more general "command context" that includes the redirs
// deciding against it for now in favour of "doing the simplest thing possible"
type builtin func([]string, redirections)

var builtins = make(map[string]builtin, 5)

func main() {
	builtins["echo"] = echo
	builtins["exit"] = exit
	builtins["type"] = toipe
	builtins["pwd"] = pwd
	builtins["cd"] = cd

	for {
		// TODO: we should have a "last command (parsing/)execution took n milliseconds metric"
		// And maybe show it in debug mode by default?

		Terminal.RawVegan()

		fmt.Print(GetPS1())
		// maybe we don't delimit by \n here? Is this baking in the assumption that every line is a new inputCommand?
		// inputCommand, err := bufio.NewReader(os.Stdin).ReadString('\n')
		inputCommand, err := Readline.GetLine()

		if err != nil {
			fmt.Println("Could not read input from stdin")
			os.Exit(commandUsageError)
		} else {
			var outputCommandFields []string
			inputCommandRunes := []rune(inputCommand)
			outputCommandBeingBuilt := ""
			commandRedirections := redirections{in: os.Stdin, out: os.Stdout, err: os.Stderr}

			for len(inputCommand) > 0 {
				startToken := GetNextStartToken(inputCommandRunes)
				DbgPrintTokenln("our new startToken", startToken, inputCommandRunes[startToken.GetPosition()])
				var endToken Token
				var redirectToken RedirectToken
				processingRedirection := false
				inputCommandRunes = inputCommandRunes[startToken.GetPosition():]
				inputCommand = inputCommand[startToken.GetPosition():]

				switch startToken.GetType() {
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

				case Redirection:
					processingRedirection = true
					redirectToken, err = GetNextRedirectTokenEnd(inputCommandRunes)
					DbgPrintf("Got RedirectEndToken: %v\n", redirectToken)
					if err != nil {
						fmt.Printf("Error while getting RedirectToken End Token: %s", err)
						os.Exit(generalError)
					}
					endToken = redirectToken.Token
					switch redirectToken.Direction {
					case RedirectOutput:
						var outputFile *os.File
						if redirectToken.ShouldAppend {
							outputFile, err = os.OpenFile(redirectToken.FileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
							if err != nil {
								fmt.Printf("Output file: '%s' for command redirect could not be opened: %s\n", redirectToken.FileName, err)
								os.Exit(IOError)
							}
							commandRedirections.out = outputFile
						} else { // should not append
							outputFile, err = os.Create(redirectToken.FileName)
							if err != nil {
								fmt.Printf("Output file: '%s' for command redirect could not be opened: %s\n", redirectToken.FileName, err)
								os.Exit(IOError)
							}
							commandRedirections.out = outputFile
						}
					case RedirectInput:
						inputFile, err := os.Open(redirectToken.FileName)
						if err != nil {
							fmt.Printf("Input file for command redirect could not be opened: %s\n", err)
							os.Exit(IOError)
						}
						commandRedirections.in = inputFile
					case RedirectError:
						var errorOutput *os.File
						if redirectToken.ShouldAppend {
							errorOutput, err = os.OpenFile(redirectToken.FileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
							if err != nil {
								fmt.Printf("Error file: '%s' for command redirect could not be opened: %s\n", redirectToken.FileName, err)
								os.Exit(IOError)
							}
							commandRedirections.err = errorOutput
						} else { // should not append
							errorOutput, err = os.Create(redirectToken.FileName)
							if err != nil {
								fmt.Printf("Error file: '%s' for command redirect could not be opened: %s\n", redirectToken.FileName, err)
								os.Exit(IOError)
							}
							commandRedirections.err = errorOutput
						}
					default:
						panic(fmt.Sprintf("unexpected main.RedirectType: %#v", redirectToken.Direction))
					}

				case Termination:
					endToken, err = Token{Position: 1, Type: Termination}, nil
					DbgPrintTokenln("our new endToken", endToken, inputCommandRunes[endToken.Position-1]) // termination is a special boy

				default:
					panic("unimplemented token type")
				}
				if processingRedirection {
				} else {
					outputCommandBeingBuilt += GetSanitisedCommandSegment(inputCommand, endToken)
				}

				inputCommand = inputCommand[endToken.Position:]
				inputCommandRunes = inputCommandRunes[endToken.Position:]
				DbgSanitisedPrintf("new command: %v\n", inputCommand)
				DbgPrintf("new commandRunes: %v\n", inputCommandRunes)

				if (!processingRedirection) && (endToken.Type != Termination) && (inputCommandRunes[0] == ' ' || inputCommandRunes[0] == '\n') {
					outputCommandFields = append(outputCommandFields, outputCommandBeingBuilt)
					DbgPrintf("new commandFields: %v\n", outputCommandFields)
					outputCommandBeingBuilt = ""
				}
			}

			DbgPrintf("We are done, done. Nothing else to process.\n")
			outputCommandName := outputCommandFields[0]

			// is this a builtin?
			if _builtin, ok := builtins[outputCommandName]; ok {
				_builtin(outputCommandFields[1:], commandRedirections)
				continue
			}

			// maybe we can run it?
			if _, exists := executableExistsInPath(outputCommandName); exists {
				cmd := exec.Command(outputCommandName, outputCommandFields[1:]...)
				DbgPrintf("running command %s with args %v\n", outputCommandName, outputCommandFields[1:])
				cmd.Stdin = commandRedirections.in
				cmd.Stdout = commandRedirections.out
				cmd.Stderr = commandRedirections.err
				cmd.Run()
				continue

			}

			// looks like we don't know what this is
			fmt.Printf("%v: command not found\n", strings.Join(outputCommandFields, ""))
		}
	}
}

func echo(params []string, redirs redirections) {
	// fmt.Println(strings.Join(params, " "))
	fmt.Fprintln(redirs.out, strings.Join(params, " "))
}

func cd(params []string, _ redirections) {
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

func pwd(params []string, redirs redirections) {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(generalError)
	}
	fmt.Fprintln(redirs.out, wd)
}

func exit(code []string, _ redirections) {
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

func toipe(fns []string, redirs redirections) {
	for _, t := range fns {
		isBuiltin, isInPath := false, false
		if _, ok := builtins[t]; ok {
			fmt.Fprintf(redirs.out, "%s is a shell builtin\n", t)
			isBuiltin = true
			continue
		}
		if fullFilePath, exists := executableExistsInPath(t); exists {
			fmt.Fprintf(redirs.out, "%s is %s\n", t, fullFilePath)
			isInPath = true
			continue
		}
		if !isBuiltin && !isInPath {
			fmt.Fprintf(redirs.out, "%s: not found\n", t)
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
