package main

import (
	"os"

	"golang.org/x/sys/unix"
)

// NOTE:
// we could have used golang.org/x/term for this, but I wanted to implement parts of the term functionality myself
// so I can better understand what's going on and which parts of the term/tty stack we are actually directly interacting with
// The makeRaw function on go/x/term (source) https://cs.opensource.google/go/x/term/+/master:term_unix.go;l=22 shows the proper way of doing this.
// I suspect I might want to ditch my own simplistic implementation and replace it with that at some point
//
// Another point of documentation is "man 2 ioctl_tty" for TCGETS and TCSETS and all the other term controls
// I was dumb when looking for that, and it took me ages to figure out that I couldn't get the to the docs with "man 2 TCGETS" so I'll leave this here as well
// If mood strikes to click on hyperlinks, this also works: https://man7.org/linux/man-pages/man2/TCGETS2.2const.html

type terminal struct{}

var Terminal terminal

// the setRaw and setCooked functions here do not save, return or restore previous term state
// Since we only mess with ICANON and ECHO, I want to see if we can get away with just toggling these on an off
// or maybe later on possibly going the zsh route and restoring the term to a sane state after each command run

func (term terminal) RawVegan() {
	stdinFD := int(os.Stdin.Fd())                          // file descriptor is uintptr, convert to int for GetTermios
	term_, _ := unix.IoctlGetTermios(stdinFD, unix.TCGETS) // Terminal Control GET Settings
	// Lflag for Line-discipline flags
	// ICANON for canonical or "cooked" mode which buffers a full line of input
	// and gives it to the shell once the line delimiter is encountered.
	// we disable this because we want to process input char by char
	// so we can process things like '\t'
	// ECHO for echoing the characters written back to the term
	// rn, we won't be able to process controll characters like ^C
	// becaues we are not including unix.ISIG (signals) or unix.IEXTEN (^V:paste)
	term_.Lflag &^= unix.ICANON | unix.ECHO
	unix.IoctlSetTermios(stdinFD, unix.TCSETS, term_)
}

func (term terminal) Cookify() {
	stdinFd := int(os.Stdin.Fd())
	term_, _ := unix.IoctlGetTermios(stdinFd, unix.TCGETS)
	term_.Lflag |= unix.ICANON | unix.ECHO
	unix.IoctlSetTermios(stdinFd, unix.TCSETS, term_)
}
