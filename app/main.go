package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
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

// considered generalizing this into a more general "command context" that includes the redirs
// deciding against it for now in favour of "doing the simplest thing possible"
type redirections struct {
	in  *os.File
	out *os.File
	err *os.File
}

type builtin func([]string, redirections)

var builtins map[string]builtin

func init() {
	builtins = map[string]builtin{
		"echo": echo,
		"exit": exit,
		"type": toipe,
		"pwd":  pwd,
		"cd":   cd,
	}
	for key := range builtins {
		Readline.Completions = append(Readline.Completions, key)
	}
}

/*
func __SimpleExampleForPipeImpl() {
	cmd := exec.Command("ls", "-lash")
	cmd2 := exec.Command("grep", "\\s\\.")
	so, _ := os.Create("dots")
	r, w, _ := os.Pipe()
	cmd.Stdout = w
	cmd2.Stdin = r
	cmd2.Stdout = so

	var wgPip sync.WaitGroup

	wgPip.Go(func() {
		if x := cmd2.Start(); x != nil {
			fmt.Printf("cmd2: err:%v\n", x)
		}
		cmd2.Wait()
		so.Close()
		r.Close()
	})
	wgPip.Go(func() {
		if y := cmd.Start(); y != nil {
			fmt.Printf("cmd: err:%v\n", y)
		}
		cmd.Wait()
		w.Close()
	})

	wgPip.Wait()
	os.Exit(0)
}
*/

func main() {
	commandHasPipes := false
	var secondCommand string // very bad! assumes only one pipe!
	var secondCommandArgs string
	for {
		// TODO: we should have a "last command (parsing/)execution took n milliseconds metric"
		// And maybe show it in debug mode by default?

		Terminal.RawVegan() // put term into raw mode

		fmt.Print(GetPS1())

		inputCommand, err := Readline.GetLine()

		// FIXME: this is hopefully the simple possible (albeit very wrong!) implementation of the pipe.
		// done as just a PoC to be replaced later with one that actually works Readline
		// hopefully this wont be the combo breaker when it comes to actually removing "temp" code
		//

		commands := strings.Split(inputCommand, "|")
		if len(commands) > 1 {
			// we will only consider one piping in this PoC
			DbgPrintf("before pipe: %s\n", commands[0])
			DbgPrintf("after pipe: '%s'\n", commands[1])
			commands[0] = strings.TrimRight(commands[0], " \n") // lets not trim left IN CASE we want to implement the history feature where commands beginning with space are not appendend
			commands[1] = strings.Trim(commands[1], " ")
			commandHasPipes = true
			endOfFirstWord := strings.Index(commands[1], " ") // find first space after the pipe "wc<spc>-l"
			if endOfFirstWord != -1 {                         // if there is an argument to the second command
				secondCommandArgs = strings.Trim(commands[1][endOfFirstWord:], " \n")
				secondCommand = commands[1][:endOfFirstWord]
			} else {
				secondCommand = strings.Trim(commands[1], " \n")
			}

			DbgPrintf("second command:      '%s'\n", secondCommand)
			DbgPrintf("second command args: '%s'\n", secondCommandArgs)

			// os.Exit(0)
			inputCommand = commands[0] + "\n" //
		}

		Terminal.Cookify() // revert changes we did for raw mode
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not read input from stdin %s \n", err)
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
				if commandHasPipes {

					r, w, _ := os.Pipe()
					var cmd2 *exec.Cmd
					secondCommandArgsSlice := strings.FieldsFunc(secondCommandArgs, func(r rune) bool { return r == ' ' })
					cmd2 = exec.Command(secondCommand, secondCommandArgsSlice...)
					cmd.Stdout = w
					cmd2.Stdin = r
					cmd2.Stdout = os.Stdout
					cmd2.Stderr = os.Stdout

					var cmdsWG sync.WaitGroup

					cmdsWG.Add(1)
					go func(theWG *sync.WaitGroup) {
						defer theWG.Done()
						defer w.Close()

						c1e := cmd.Run()
						if c1e != nil {
							DbgPrintf("cmd error: %s\n", c1e)
						}
					}(&cmdsWG)

					cmdsWG.Add(1)
					go func(theWG *sync.WaitGroup) {
						defer theWG.Done()
						defer r.Close()

						c2e := cmd2.Run()
						if c2e != nil {
							DbgPrintf("cmd2 error: %s\n", c2e)
						}
					}(&cmdsWG)

					cmdsWG.Wait()

				} else {
					cmd.Stdin = commandRedirections.in
					cmd.Stdout = commandRedirections.out
					cmd.Stderr = commandRedirections.err
					cmd.Run()
				}
				commandHasPipes = false
				secondCommand = ""
				secondCommandArgs = ""
				continue

			}

			// looks like we don't know what this is
			// fmt.Fprintf(commandRedirections.out, "%s: command not found\n", strings.Join(outputCommandFields, ""))
			fmt.Fprintf(commandRedirections.out, "%s: command not found\n", outputCommandName)
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
