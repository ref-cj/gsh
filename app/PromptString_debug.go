//go:build debug
// +build debug

package main

import (
	"os"
	"os/user"
	"strings"
)

// putting this here mainly to preserve the colours for future reference
// stuff is hard-coded right now so it doesn't have much use other than aesthetics
// colours gleefully stolen from Quantum Terminal example @ https://colormyshell.tcpip.wtf/
const (
	reset          = "\033[0m"
	clrBookEnd     = "\033[38;2;191;97;106m"
	clrPunctuation = "\033[1;38;2;180;142;173m"
	clrShell       = "\033[38;2;235;203;139m"
	clrData1       = "\033[38;2;163;190;140m"
	clrData2       = "\033[38;2;94;129;172m"
	clrConnector   = "\033[38;2;129;161;193m"
	space          = " "
)
const numberOfPathSegmentsToShow = 2

var (
	userName string
	hostName string
)

func init() {
	currentUser, err := user.Current()
	if err != nil {
		panic("couldn't get user")
	}
	userName = currentUser.Username
	hostName, err = os.Hostname()
	if err != nil {
		panic("could not get hostname")
	}
}

//TODO: This is the draft of a draft of a plan for implementation. Recording it here mostly so I don't have to start thinking about it from 0 the next time
// I want to add a "last command took: XXms" part to this. But this PS is getting a bit crowded.
// We can split this into two parts and put some of the info to the right.
// The timing info we can take in as a parameter or like a context containing timing info and other relevant context.
// This might include the username and such. Arguably it should be someonelse's responsibility to get, cache and invalidate it.
// After we have the context, we can do what go/x/term does (https://cs.opensource.google/go/x/term/+/master:term_unix.go)
// and get the width with unix.IoctlGetWinsize(fd, unix.TIOCGWINSZ).Column.
// then measure the left frompt string, and right-prompt-string; figure out if there is enough space
// if there is, render right-prompt, if not, either frop it like zsh, or maybe put it on the next line?
//
// More TODO: if every new comman has a timing associated with it, we should probably also keep a history to go back and check
// maybe wait for the actual history implementation for this?
//

func GetPS1() string {
	// NOTE: caching this and invalidating it in the `cd` builtin is attractive
	// but when measured, this added 10µs. Maybe not worth the complexity and potential bugs?
	pathSegmentToShow := getPathSegmentToShow()
	return space + clrBookEnd + "🙝" + reset + space + space +
		clrPunctuation + "[" + reset + clrShell + "gsh" + reset + clrPunctuation + "]" + reset + space +
		clrData1 + userName + reset + clrPunctuation + "@" + reset + clrData1 + hostName + reset + space +
		clrConnector + "→" + reset + space + clrData2 + pathSegmentToShow + reset + space +
		clrBookEnd + "🙞" + reset + space + space
}

func getPathSegmentToShow() string {
	currentDir, _ := os.Getwd()
	// NOTE: this makes a possibly big string slice when the path is deep
	// once we are happy with the amount of path segment we want to show
	// maybe consider converting this to a loop that goes from the end of the path and counts slashes.
	pathSegments := strings.Split(currentDir, string(os.PathSeparator))
	var shownPath string
	if len(pathSegments) > numberOfPathSegmentsToShow {
		shownPath += "⋯/"
	}
	shownPath += strings.Join(pathSegments[len(pathSegments)-numberOfPathSegmentsToShow:], "/")
	return shownPath
}
