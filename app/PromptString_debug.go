//go:build debug
// +build debug

package main

import (
	"os"
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

func GetPS1() string {
	// NOTE: caching this and invalidating it in the `cd` builtin is attractive
	// but when measured, this added 10µs. Maybe not worth the complexity and potential bugs?
	shownPath := getPathSegmentToShow()

	return space + clrBookEnd + "🙝" + reset + space + space +
		clrPunctuation + "[" + reset + clrShell + "gsh" + reset + clrPunctuation + "]" + reset + space +
		clrData1 + "user" + reset + clrPunctuation + "@" + reset + clrData1 + "host" + reset + space +
		clrConnector + "→" + reset + space + clrData2 + shownPath + reset + space +
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
