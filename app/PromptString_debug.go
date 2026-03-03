//go:build debug
// +build debug

package main

// putting this here mainly to preserve the colours for future reference
// stuff is hard-coded right now so it doesn't have much use other than aesthetics
// gleefully stolen from Quantum Terminal example @ https://colormyshell.tcpip.wtf/
const (
	reset          = "\033[0m"
	clrBookEnd     = "\033[38;2;191;97;106m"
	clrPunctuation = "\033[1;38;2;180;142;173m"
	clrShell       = "\033[38;2;235;203;139m"
	clrData1       = "\033[38;2;163;190;140m"
	clrData2       = "\033[38;2;94;129;172m"
	clrConnector   = "\033[38;2;129;161;193m"
	space          = " "
	prompt         = space + clrBookEnd + " 🙝" + reset + space + space +
		clrPunctuation + "[" + reset + clrShell + "gsh" + reset + clrPunctuation + "]" + reset + space +
		clrData1 + "user" + reset + clrPunctuation + "@" + reset + clrData1 + "host" + reset + space +
		clrConnector + "→" + reset + space + clrData2 + "~/current/path" + reset + space +
		clrBookEnd + "🙞" + reset + space + space
)
